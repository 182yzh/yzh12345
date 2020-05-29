package flowscheduler

const (
	TASK_UNSCHEDULED = 1
	TASK_RUNNING = 2
	TASK_FAILED = 3
	TASK_COMPLETED = 4 
)

type TaskDescriptor struct{
	Tid TaskID
	Name string
	State TaskState
	ResRequest ResVector
	Jd  *JobDescriptor
	Priority uint64
	
	SubmitTime uint64
	StartTime uint64
	FinishTime uint64

	//Labels []Label
	//LabelSelector []LabelSelector
}

func (td *TaskDescriptor)GetTaskID()TaskID{
	return td.Tid
}

func (td *TaskDescriptor)GetJobID()JobID{
	return td.Jd.GetJobID()
}



