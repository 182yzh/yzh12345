package flowscheduler

type NodeID = uint64
type NodeType = uint8
type TaskID = uint64
type TaskType = uint8
type JobID = uint64
type ResID = uint64
type JobState = uint8
type ResType = uint8
type TaskState = uint8
type ArcType = uint8

type ResVector struct {
	Cpu uint64
	Memory uint64
	Gpu uint64
}

type Label struct{
	key string
	value string
}

type ArcDescriptor struct {
	capLower uint64
	capUpper uint64
	cost     uint64
}

func Test()string {
	return "test \n"
}