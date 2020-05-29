package flowscheduler

import "fmt"
import "time"
import "vlog"

type FlowScheduler struct{
	gm      *GraphManager
	dp      *Dispatcher
	cm      CostModel
	// the pus that was removed during the slover run
	//pusRemoved       map[NodeID]bool
	//taskCompleted    map[NodeID]bool
	jobs      		 map[JobID]*JobDescriptor
	toScheduleJobs   []*JobDescriptor
	tasks            map[TaskID]*TaskDescriptor
	ress             map[ResID]*ResDescriptor
	taskBindings    map[TaskID]ResID
	resBindings     map[ResID](map[TaskID]bool)
	runnableTask     map[JobID](map[TaskID]bool)
	lastUpdateTime   uint64
	//if time has passed updateFrequncy(ms) ,then we update time based cost
	UpdateFrequency  uint64
	sloverRunCnt     uint64
}


func NewFlowScheduler()*FlowScheduler{
	fs := new(FlowScheduler)
	fs.gm = new(GraphManager)
	fs.gm.Init()
	fs.dp = NewDispatcher(fs.gm)
	fs.cm = NewGpuCostModel(fs.gm)
	//fs.pusRemoved = make(map[NodeID]bool)
	//fs.taskCompleted = make(map[NodeID]bool)
	fs.jobs = make(map[JobID]*JobDescriptor)
	fs.tasks = make(map[TaskID]*TaskDescriptor)
	fs.ress =make(map[ResID]*ResDescriptor)
	fs.taskBindings = make(map[TaskID]ResID)
	fs.resBindings = make(map[ResID](map[TaskID]bool))
	fs.runnableTask = make(map[JobID](map[TaskID]bool))
	fs.toScheduleJobs = make([]*JobDescriptor,0,0)
	return fs
}

func (fs *FlowScheduler)ScheduleAllJobs()map[TaskID]ResID{
	ans := fs.ScheduleJobs(fs.toScheduleJobs)
	fs.toScheduleJobs = make([]*JobDescriptor,0,0)
	return ans
}

func (fs *FlowScheduler)TaskIDToJobDes(tid TaskID)*JobDescriptor{
	td,ok := fs.tasks[tid]
	if ok == false {
		vlog.Vlog("fs.TaskIDToJobDes, td is nil")
		return nil
	}
	return td.Jd
}

func (fs *FlowScheduler)ResIDToResDes(rid ResID)*ResDescriptor{
	rd,ok := fs.ress[rid]
	if !ok {
		return nil
	}
	return rd
}
func (fs *FlowScheduler)ScheduleJobs(jobs []*JobDescriptor)map[TaskID]ResID {
	fs.gm.AddOrUpdateJobNodes(jobs)
	return fs.RunScheduleIteration()
}




func (fs *FlowScheduler)HandleTaskCompeleted(td *TaskDescriptor){
	//if task is abort or failed ,is should be already removed
	tid := td.GetTaskID()
	delete(fs.tasks,tid)
	jid := td.Jd.GetJobID()
	delete(fs.runnableTask[jid],tid)
	if rid,ok := fs.taskBindings[tid];ok{
		delete(fs.taskBindings,tid)
		delete(fs.resBindings[rid],tid)
	}
	tnode := fs.gm.TaskIDToNode(td.GetTaskID())
	fs.gm.TaskCompleted(tnode)
	//fs.CheckJobCompleted(td.Jd)
}


func (fs *FlowScheduler)HandleTaskPlacement(tid TaskID,rid ResID){
	_,ok := fs.tasks[tid];
	if !ok {
		vlog.Dlog(fmt.Sprintf("Error-fs-HandleTaskPlacement,the task(taskID is %d)is not exists",tid))
		return
	}
	rd,ok := fs.ress[rid];
	if !ok {
		vlog.Dlog(fmt.Sprintf("Error-fs-HandleTaskPlacement,the res(resID is %d) is not exists",rid))
		return
	}
	
	rd.AddCurrentRunningTask(tid)
	fs.gm.TaskScheduled(fs.gm.TaskIDToNode(tid),fs.gm.ResIDToNode(rid))
	fs.taskBindings[tid]=rid
	if _,ok := fs.resBindings[rid];!ok{
		fs.resBindings[rid] = make(map[TaskID]bool)
	}
	fs.resBindings[rid][tid]=true
}


func (fs *FlowScheduler)RunScheduleIteration()map[TaskID]ResID{
	curTime := uint64(time.Now().Unix())
	fs.lastUpdateTime = uint64(curTime)
	/*alljobs := make([]*JobDescriptor,0,0)
	for _,jd := range fs.jobs {
		alljobs = append(alljobs,jd)
	}
	

	fs.gm.AddOrUpdateJobNodes(alljobs)
	*/
	// run slover to get the task to res mapping
	output,taskMappings := fs.dp.RunSolver()
	fmt.Sprintf(output)
	
	taskToRes := make(map[TaskID]ResID)
	for tnid,rnid := range taskMappings{
		tnode := fs.gm.gcm.GetNode(tnid)
		rnode := fs.gm.gcm.GetNode(rnid)
		if !fs.HaveEnoughResource(tnode.td,rnode.rd){
			continue
		}
		tid := tnode.td.GetTaskID()
		rid := rnode.rd.GetResID()
		fs.HandleTaskPlacement(tid,rid)
		taskToRes[tid] = rid
	}
	taskMappings = make(map[NodeID]NodeID)
	return taskToRes
}


func (fs *FlowScheduler)HaveEnoughResource(td *TaskDescriptor,rd *ResDescriptor)bool{
	rr := td.ResRequest
	ar := rd.ResAvailable
	return  rr.Gpu<=ar.Gpu  && rr.Cpu<=ar.Cpu && rr.Memory <= ar.Memory
}


func (fs *FlowScheduler)CheckJobCompleted(jd *JobDescriptor){
	if tasks,ok := fs.runnableTask[jd.GetJobID()];ok{
		if len(tasks)>0{
			return 
		}
	}
	fs.HandleJobCompleted(jd)
}

func (fs *FlowScheduler)HandleJobCompleted(jd *JobDescriptor){
	for _,td := range jd.Tasks {
		if td.State == TASK_UNSCHEDULED || td.State == TASK_RUNNING {
			fs.HandleTaskCompeleted(td)
		}
	}
	jid := jd.GetJobID()
	fs.gm.JobCompleted(jd)
	delete(fs.jobs,jid)
	delete(fs.runnableTask,jid)
}

func (fs *FlowScheduler)AddJob(jd *JobDescriptor){
	jid := jd.GetJobID()
	fs.jobs[jid] = jd
	if _,ok := fs.runnableTask[jid];!ok{
		mp := make(map[TaskID]bool)
		fs.runnableTask[jid] =  mp
	}
	for _,td := range jd.Tasks{
		fs.tasks[td.GetTaskID()] = td
		fs.runnableTask[jid][td.GetTaskID()] = true
	}
	fs.toScheduleJobs = append(fs.toScheduleJobs,jd)
}

func (fs *FlowScheduler)AddJobs(jds []*JobDescriptor){
	for _,jd := range jds{
		fs.AddJob(jd)
	}
}
func (fs *FlowScheduler)AddResource(rd *ResDescriptor){
	rid := rd.GetResID()
	fs.ress[rid] = rd
	rds := make([]*ResDescriptor,0,0)
	rds = append(rds,rd)
	fs.gm.AddOrUpdateAllResNodes(rds)
}

func (fs *FlowScheduler)ExportGraph()string{
	return fs.gm.ExportGraph()
}



