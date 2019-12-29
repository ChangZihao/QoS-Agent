package podInfo

import (
	"github.com/prometheus/common/log"
	"io/ioutil"
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

func GerAllContainer(dir string) []string {
	var dockers []string
	fileInfos, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Errorf("Read pod root dir failed!")
	}
	for _, file := range fileInfos {
		if file.IsDir() && strings.Contains(file.Name(), "docker-") {
			filepath := dir + "/" + file.Name()
			dockers = append(dockers, filepath)
		}
	}
	if len(dockers) > 0 {
		return dockers
	} else {
		return nil
	}
}

func GetPodPids(podpath string) []string {
	var pids []string
	fileInfos, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Errorf("Read pod root dir failed!")
	}
	for _, file := range fileInfos {
		if file.IsDir() && strings.Contains(file.Name(), "docker-") {
			dockerPath := podpath + "/" + file.Name()
			data, err := ioutil.ReadFile(dockerPath + "/cgroup.procs")
			if err != nil {
				log.Errorf("Get docker %s pid failed!", dockerPath)
			}
			pids = append(pids, string(data))
		}
	}
	if len(pids) > 0 {
		return pids
	} else {
		return nil
	}
}
