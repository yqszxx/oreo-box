package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {

		if res.CpuSetCpus != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSetCpus), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset.cpus fail %v", err)
			}
		} else {
			ResetValue(&res.CpuSetCpus, "0")
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSetCpus), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset.cpus fail %v", err)
			}
		}

		if res.CpuSetMems != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.mems"), []byte(res.CpuSetMems), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset.mems fail %v", err)
			}
		} else {
			ResetValue(&res.CpuSetMems, "0")
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.mems"), []byte(res.CpuSetMems), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset.mems fail %v", err)
			}
		}

		return nil
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func ResetValue(s *string, newValue string) {
	sByte := []byte(*s)
	for i := 0; i < len(sByte); i++ {
		sByte[i] = ' '
	}
	*s = newValue
}
