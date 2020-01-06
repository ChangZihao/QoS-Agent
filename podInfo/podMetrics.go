package podInfo

import (
	"bufio"
	"fmt"
	"github.com/prometheus/common/log"
	"node-exporter/collector"
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

func StarqposMonitor(pod string, m *sync.Map) *exec.Cmd {
	if podPath, isFind := GetPod(pod); isFind {
		pids := GetPodPids(podPath)
		fmt.Println(pids)
		cmd, stdout := Monitor(pids)
		go func() {
			for {
				content := utils.ReadLines(len(pids)+2, 2, stdout)
				data := utils.GetpqosFormat(content)
				if data != nil {
					m.Store(pod, data)
				} else {
					collector.PQOSMetrics.Delete(pod)
					break
				}
			}

			log.Infof("Stop Monitor %s", pod)
		}()
		return cmd
	} else {
		return nil
	}
}

func StopMonitor(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
