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

func StarqposMonitor(pod string, pidMap *sync.Map) (*exec.Cmd, []string) {
	if podPath, isFind := podInfo.GetPod(pod); isFind {
		pids := podInfo.GetPodPids(podPath)
		log.Infof("pids in pod %s: %v", pod, pids)
		cmd, stdout := Monitor(pids)
		go func() {
			for {
				content := utils.ReadLines(len(pids)+2, 2, stdout)
				data := utils.GetpqosFormat(content)
				if data != nil {
					PQOSMetrics.Store(pod, data)
				} else {
					log.Infof("%s monitor get nil response, exit!", pod)
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

func StopMonitor(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
