package main

import (
	"fmt"
	"github.com/prometheus/common/log"
	"math/bits"
	"node-exporter/collector"
	"node-exporter/podInfo"
	"node-exporter/utils"
	"os/exec"
	"strconv"
	"strings"
)

var resctrlPath = "/sys/fs/resctrl"

type LLCCos struct {
	CosId  int
	Status bool // false -> not run, true running
	LLCMap uint // bit 1 -> used
}

func (cos LLCCos) AllocCount() int {
	return bits.OnesCount(cos.LLCMap)
}

func (cos LLCCos) AllocMap() uint {
	return cos.LLCMap
}

type LLCManager struct {
	MaxCos    int
	MaxLLCWay int
	CosMap    uint // bit 0 -> available, Cos 0 is reserved
	LLCMap    uint // bit 0 -> available
	AllocInfo map[string]*LLCCos
}

func NewLLCManager() LLCManager {
	// TODO fix get cosCount & llcCount auto from the system
	manager := LLCManager{16, 20, 0, 0, make(map[string]*LLCCos)}
	manager.ResetAlloc()
	return manager
}

func (llc *LLCManager) AllocCount() int {
	return bits.OnesCount(llc.LLCMap)
}

func (llc *LLCManager) AllocLLC(pod string, value string) bool {
	if iValue, err := strconv.Atoi(value); err != nil || iValue < 0 {
		log.Infof("AllocLLC %s failed, wrong llc value %s", pod, value)
		return false
	} else {
		if _, ok := llc.AllocInfo[pod]; iValue == 0 && ok {
			llc.ReleaseCos(pod)
		} else if llc.AllocCos(pod) {
			if ok := llc.SetLLCAlloc(pod, iValue); ok == true {
				log.Infof("Set LLC alloc for %s(%d) success!", pod, iValue)
				return true
			} else {
				log.Errorf("Set LLC alloc for %s(%d) failed!", pod, iValue)
				if llc.AllocInfo[pod].Status == false {
					llc.ReleaseCos(pod)
				}
			}
		}
	}
	return false
}

func (llc *LLCManager) AllocCos(pod string) bool {
	if _, ok := llc.AllocInfo[pod]; !ok {
		for index := llc.MaxCos - 1; index > 0; index-- {
			if llc.CosMap&(1<<uint(index)) == 0 {
				llc.AllocInfo[pod] = &LLCCos{index, false, 0}
				//TODO Check add pids to cgroup
				pid := podInfo.GetPidByPodName(pod)
				pidStr := utils.StrList2lines(pid)
				cosCMD := fmt.Sprintf("echo -e \"%s\" > %s/COS%d/tasks", pidStr, resctrlPath, index)
				cmd := exec.Command("/bin/bash", "-c", cosCMD)
				out, err := cmd.CombinedOutput()
				if err != nil || len(out) > 0 {
					log.Errorf("Do %s failed! err: %s, out: %s", cosCMD, err, out)
				}
				llc.CosMap += 1 << uint(index)
				return true
			}
		}
		log.Errorf("Can not find free COS for %s", pod)
		return false
	} else {
		return true
	}
}

func (llc *LLCManager) ReleaseCos(pod string) bool {
	llc.CosMap -= 1 << llc.AllocInfo[pod].CosId
	//Todo  write pid to cos0pod
	cmdStr := fmt.Sprintf("cat /sys/fs/resctrl/COS%d/tasks | xargs -I {} echo {} > /sys/fs/resctrl/tasks", llc.AllocInfo[pod].CosId)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) > 0 {
		log.Errorf("Do %s failed! err: %s, our:%s", cmdStr, err, out)
	}
	delete(llc.AllocInfo, pod)
	log.Infof("Release cos for %s success", pod)
	llc.LLCMap = llc.SumCosLLCAlloc()
	return true
}

func (llc *LLCManager) SetLLCAlloc(pod string, value int) bool {
	mask := llc.AllocLLCMask(pod, value) //mask == 0, failed
	//Todo check write cgroup
	if mask == 0 {
		log.Errorf("Can not find suitable LLC, map: 0x%b", llc.LLCMap)
		return false
	}
	maskFile := fmt.Sprintf("%s/COS%d/schemata", resctrlPath, llc.AllocInfo[pod].CosId)
	maskStr := utils.UInt2BitsStr(mask, llc.MaxLLCWay)
	maskCMD := fmt.Sprintf("echo \"L3:0=%s;1=%s\" > %s", maskStr, maskStr, maskFile)
	cmd := exec.Command("/bin/bash", "-c", maskCMD)
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) > 0 {
		log.Errorf("Do %s failed! err: %s, out: %s", maskCMD, err, out)
		return false
	} else {
		llc.AllocInfo[pod].LLCMap = mask
		llc.AllocInfo[pod].Status = true
		llc.LLCMap = llc.SumCosLLCAlloc()
		collector.LLCAllocCount.Store(pod, value)
		return true
	}
}

func (llc *LLCManager) AllocLLCMask(pod string, value int) uint {
	curLLCMap := llc.LLCMap
	if llc.AllocInfo[pod].Status == true {
		curLLCMap -= llc.AllocInfo[pod].LLCMap
	}
	for mask := uint((1 << value) - 1); mask < (1 << llc.MaxLLCWay); mask = mask << 1 {
		if mask&curLLCMap == 0 {
			return mask
		}
	}
	return 0
}

func (llc *LLCManager) SumCosLLCAlloc() uint {
	total := uint(0)
	for _, v := range llc.AllocInfo {
		total += v.LLCMap
	}
	return total
}

func (llc *LLCManager) ResetAlloc() bool {
	cmd := exec.Command("pqos", "-I", "-R")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorln("cmd StdoutPipe error:", err)
		return false
	}

	result := string(out)
	if err != nil || !strings.Contains(result, "successful") {
		log.Errorf("LLC alloc reset fail！%s. err: %s", result, err)
		return false
	}
	for i := 1; i < 16; i++ {
		cmdStr := fmt.Sprintf("cat /sys/fs/resctrl/COS%d/tasks | xargs -I {} echo {} > /sys/fs/resctrl/tasks", i)
		cmd = exec.Command("/bin/bash", "-c", cmdStr)
		cmd.Run()
	}

	log.Infof("Reset llc alloc success!")
	return true
}
