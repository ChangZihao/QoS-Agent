package podInfo

import (
	"fmt"
	"io"
	"os/exec"
)

func monitor(pids []string) (int, io.ReadCloser) {
	if pids != nil {
		fmt.Println("Pid is nil in monitor!")
		return -1, nil
	}
	paraPid := "all:"
	for _, pid := range pids {
		paraPid += pid + ","
	}
	cmd := exec.Command("pqos", "-I", "-p", paraPid)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("cmd StdoutPipe error:", err)
		return -1, nil
	}
	go cmd.Start()
	return cmd.Process.Pid, stdout
}
