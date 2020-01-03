package podInfo

import (
	"bufio"
	"github.com/prometheus/common/log"
	"node-exporter/utils"
	"os/exec"
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

func StopMonitor(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
