package subsystems

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSetCpus  string
	CpuSetMems  string
	CpuQuotaUs  string
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
