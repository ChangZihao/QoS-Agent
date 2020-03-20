package main

import (
	"github.com/prometheus/common/log"
	"math/bits"
	"os/exec"
	"strconv"
	"strings"
)

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
	//manager.ResetAlloc()
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
		if llc.AllocCos(pod) {
			if ok := llc.SetLLCAlloc(pod, iValue); ok == true {
				log.Infof("Set LLC alloc for %s(%d) success!", pod, iValue)
				return true
			} else {
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
				//TODO add pids to cgroup
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

func (llc *LLCManager) TempReleaseLLC(pod string) bool {
	//just temp drop llc alloc info from global llc alloc info
	if cos, ok := llc.AllocInfo[pod]; ok {
		llc.LLCMap = llc.LLCMap - cos.LLCMap
	} else {
		log.Errorf("Alloc info not found for %s", pod)
		return false
	}
	return true
}

func (llc *LLCManager) ReleaseCos(pod string) bool {
	llc.CosMap -= 1 << llc.AllocInfo[pod].CosId
	//Todo  write pid to cos0
	delete(llc.AllocInfo, pod)
	return true
}

func (llc *LLCManager) SetLLCAlloc(pod string, value int) bool {
	mask := llc.AllocLLCMask(pod, value) //mask == 0, failed
	//Todo  write cgroup
	ok := true // ,,,,,

	if ok && mask != 0{
		llc.AllocInfo[pod].LLCMap = mask
		llc.AllocInfo[pod].Status = true
		llc.LLCMap = llc.SumCosLLCAlloc()
		return true
	}
	return false
}

func (llc *LLCManager) AllocLLCMask(pod string, value int) uint {
	curLLCMap := llc.LLCMap
	if llc.AllocInfo[pod].Status == true {
		curLLCMap -= llc.AllocInfo[pod].LLCMap
	}
	for mask := uint((1 << value) - 1); mask < (1 << llc.MaxLLCWay); mask = mask << 1 {
		if mask & curLLCMap == 0 {
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
	cmd.Run()
	result := string(out)
	if !strings.Contains(result, "successful") {
		log.Errorf("LLC alloc reset fail！%s", result)
		return false
	}
	log.Infof("Reset llc alloc success! %s", result)
	return true
}
