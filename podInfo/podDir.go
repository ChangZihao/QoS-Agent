package podInfo

import (
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"strconv"
	"strings"
)

var (
	rootDir = "/sys/fs/cgroup/cpu/kubepods.slice/kubepods-burstable.slice"
)

func GetAllPod() []string {
	var pods []string
	fileInfos, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Error("Read pod root dir failed!")
	}
	for _, file := range fileInfos {
		if file.IsDir() && strings.Contains(file.Name(), "kubepods-") {
			filepath := rootDir + "/" + file.Name()
			pods = append(pods, filepath)
		}
	}
	if len(pods) > 0 {
		return pods
	} else {
		return nil
	}
}

func GetPod(name string) (string, bool) {
	pods := GetAllPod()
	for _, pod := range pods {
		if strings.Contains(pod, name) {
			return pod, true
		}
	}
	return "", false
}

func GetAllContainer(podPath string) []string {
	var dockers []string
	fileInfos, err := ioutil.ReadDir(podPath)
	if err != nil {
		log.Errorf("Read pod root podPath failed!")
	}
	for _, file := range fileInfos {
		if file.IsDir() && strings.Contains(file.Name(), "docker-") {
			filepath := podPath + "/" + file.Name()
			dockers = append(dockers, filepath)
		}
	}
	if len(dockers) > 0 {
		return dockers
	} else {
		return nil
	}
}

func GetPodPids(podPath string) []string {
	var pids []string
	fileInfos, err := ioutil.ReadDir(podPath)
	if err != nil {
		log.Errorf("Read pod root dir failed!")
	}
	for _, file := range fileInfos {
		if file.IsDir() && strings.Contains(file.Name(), "docker-") {
			dockerPath := podPath + "/" + file.Name()
			data, err := ioutil.ReadFile(dockerPath + "/cgroup.procs")
			if err != nil {
				log.Errorf("Get docker %s pid failed!", dockerPath)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if len(line) > 0 {
					pids = append(pids, strings.TrimSpace(line))
				}
			}
		}
	}
	if len(pids) > 0 {
		return pids
	} else {
		return nil
	}
}

func GetPodCPUPath(pod string) string {
	path := fmt.Sprintf("%s/kubepods-burstable-%s.slice", rootDir, pod)
	return path
}

func GetPodCPUShare(pod string) float64 {
	podPath := GetPodCPUPath(pod)
	data, err := ioutil.ReadFile(podPath + "/cpu.shares")
	if err != nil {
		log.Errorf("Get pod  %s CPUShare failed! err:%s", pod, err)
	}
	sData := strings.TrimSpace(string(data))
	if share, err := strconv.ParseFloat(sData, 64); err != nil {
		log.Errorf("CPUShare %s to float64 failed! err:%s", string(data), err)
		return 0
	} else {
		return share
	}
}

func SetPodCPUShare(pod string, value string) bool {
	shareFilePath := GetPodCPUPath(pod) + "/cpu.shares"
	if err := ioutil.WriteFile(shareFilePath, []byte(value), 0644); err != nil {
		log.Errorf("Set %s CPUShare %s failed! err:%s", pod, value, err)
		return false
	}
	return true
}
