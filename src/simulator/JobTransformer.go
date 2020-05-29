package simulator

import (
	"time"
	"flowscheduler"
	//""
)

type JobTransFormer struct {
	curjid uint64
	unusedJobid map[uint64]bool
	curtid uint64
	unusedTaskid map[uint64]bool
}

func (jtf *JobTransFormer)Init(){
	jtf.unusedJobid = make(map[uint64]bool)
	jtf.unusedTaskid = make(map[uint64]bool)
}
func (jtf *JobTransFormer)GetNextJobID()uint64{
	var ans uint64
	if len(jtf.unusedJobid)  == 0 {
		jtf.curjid++
		ans = jtf.curjid
		return ans
	}
	for id,_ := range jtf.unusedJobid{
		ans = id
		break
	}
	delete(jtf.unusedJobid,ans)
	return ans
}

func (jtf *JobTransFormer)GetNextTaskID()uint64{
	var ans uint64
	if len(jtf.unusedTaskid)  == 0 {
		jtf.curtid++
		ans = jtf.curtid
		return ans
	}
	for id,_ := range jtf.unusedTaskid{
		ans = id
		break
	}
	delete(jtf.unusedTaskid,ans)
	return ans
}


func (jtf *JobTransFormer)Transform(jinfo *JobInfo)*flowscheduler.JobDescriptor{
	jd := new(flowscheduler.JobDescriptor)
	jd.Name = jinfo.Jobid
	jd.Jid = flowscheduler.JobID(jtf.GetNextJobID())
	jd.SubmitTime = uint64(time.Now().UnixNano())
	jd.State = flowscheduler.JOB_CREATED
	td := new(flowscheduler.TaskDescriptor)
	jd.Tasks = make([]*flowscheduler.TaskDescriptor,0,0)
	jd.Tasks  = append(jd.Tasks,td)
	td.Tid = flowscheduler.TaskID(jtf.GetNextTaskID())
	gpu := jinfo.GetGpuNeedNum()
	td.Name = jd.Name+"_task"
	td.State = flowscheduler.TASK_UNSCHEDULED
	td.SubmitTime = uint64(time.Now().UnixNano())
	td.Jd = jd
	td.ResRequest.Gpu = gpu
	return jd
}

func (jtf JobTransFormer)TransFormJobs(jinfos []*JobInfo)[]*flowscheduler.JobDescriptor{
	ans := make([]*flowscheduler.JobDescriptor,0,0)
	for _,jinfo := range jinfos{
		ans = append(ans,jtf.Transform(jinfo))
	} 
	return ans
}