package collector

import (
	"bufio"
	"github.com/prometheus/common/log"
	"node-exporter/podInfo"
	"node-exporter/utils"
	"os/exec"
	"sync"
	"syscall"
)

// 构建pod pqos监控命令
func Monitor(pids []string) (*exec.Cmd, *bufio.Reader) {
	if pids == nil {
		log.Errorln("Pid is nil in monitor!")
		return nil, nil
	}
	paraPid := "all:"
	for _, pid := range pids {
		paraPid += pid + ","
	}
	cmd := exec.Command("pqos", "-I", "-p", paraPid)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorln("cmd StdoutPipe error:", err)
		return nil, nil
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	utils.SkipHead(2, reader)
	return cmd, reader
}

//启动pqos监控
func StarqposMonitor(pod string, pidMap *sync.Map) (*exec.Cmd, []string) {
	if podPath, isFind := podInfo.GetPod(pod); isFind {
		pids := podInfo.GetPodPids(podPath)
		log.Infof("pids in pod %s: %v", pod, pids)
		cmd, stdout := Monitor(pids)
		// 通过协程不断读取pqos输出
		go func() {
			for {
				content := utils.ReadLines(len(pids)+2, 2, stdout)
				data := utils.GetpqosFormat(content)
				// pqos正常返回，保存数据
				if data != nil {
					PQOSMetrics.Store(pod, data)
				// pqos 返回nil，停止监控
				} else {
					log.Infof("%s monitor get nil response, exit!", pod)
					// 删除对应数据
					PQOSMetrics.Delete(pod)
					MonitorCMD.Delete(pod)
					Pod2app.Delete(pod)
					pidMap.Delete(pod)
					break
				}
			}
			log.Infof("Stop Monitor %s", pod)
		}()
		return cmd, pids
	} else {
		return nil, nil
	}
}

// 停止监控
func StopMonitor(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
