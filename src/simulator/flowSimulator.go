package simulator

import "fmt"
import "time"
import "vlog"
import "os"
import "container/heap"
import "flowscheduler"
import "container/list"

const (
	SUBMITTED = 1
	COMPLETED = 2
	LOGTIME = 3
)

type SimInfo struct{
	curTime string
	jinfo *JobInfo
	kind    uint8
}

type SimHeap []*SimInfo
func (sh SimHeap)Len()int{return len(sh)}
func (sh SimHeap)Swap(i,j int){sh[i],sh[j] = sh[j],sh[i]}
func (sh SimHeap)Less(i,j int)bool{
	itime,err := time.ParseInLocation("2006-01-02 15:04:05",sh[i].curTime,time.Local)
	if err != nil {
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}
	jtime,err := time.ParseInLocation("2006-01-02 15:04:05",sh[j].curTime,time.Local)
	if err != nil{
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}

	ans := itime.Sub(jtime)
	return ans < 0 
}

func cmp(isTime,jsTime string)int64{
	itime,err := time.ParseInLocation("2006-01-02 15:04:05",isTime,time.Local)
	if err != nil {
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}
	jtime,err := time.ParseInLocation("2006-01-02 15:04:05",jsTime,time.Local)
	if err != nil{
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}

	ans := itime.Sub(jtime)
	//vlog.Dlog(fmt.Sprintln(isTime,jsTime,ans,int64(ans.Seconds())))
	return int64(ans.Seconds()) 
}
	

func (sh *SimHeap)Push(x interface{}) {
	*sh = append(*sh, x.(*SimInfo))
}

func (sh *SimHeap) Pop() interface{} {
	old := *sh
	n := len(old)
	item := old[n-1]
	*sh = old[0 : n-1]
	return item
}
type FlowSimulator struct {
	sh *SimHeap
	logfile *os.File
	wtlog *os.File // waiting time
	mulog *os.File //machine utilization
	reslog,qlenlog *os.File // res and queue len
	parser *JobParser
	jtf *JobTransFormer
	jinfos []*JobInfo
	fs *flowscheduler.FlowScheduler
	rg *ResTransformer
	desToInfo map[*flowscheduler.JobDescriptor]*JobInfo
	infoToDes map[*JobInfo]*flowscheduler.JobDescriptor
	jobNum int
	ResTotal,ResAvailable flowscheduler.ResVector
	queue *list.List
	timePer float64
}


func (sim *FlowSimulator)FlowWaitingTimeLog(curtime string ,jinfo *JobInfo){
	if cmp(jinfo.Submitted_time,"2017-10-03 00:00:00") < 0{
		return 
	}
	if cmp(jinfo.Submitted_time,"2017-12-10 00:00:00") > 0{
		return
	} 
	if cmp(jinfo.Attempts[0].Start_time,"2017-12-10 00:00:00") > 0{
		return
	}
	wtime := cmp(curtime,jinfo.Submitted_time)
	sim.wtlog.Write([]byte(fmt.Sprintf("%d %d\n",jinfo.GetGpuNeedNum(),wtime)))
}


func (sim *FlowSimulator)FlowMachineUtiLog(rid uint64,curtime string,jinfo *JobInfo){
	startTime := curtime
	//vlog.Vlog(fmt.Sprintln(curtime,jinfo))
	JobExcuteTime := jinfo.ExcuteTime()
	TaskStartTime,err := time.ParseInLocation("2006-01-02 15:04:05",curtime,time.Local)
	if err != nil {
		fmt.Println(err)
	}
	
	gpunum := jinfo.GetGpuNeedNum()
	
	TaskEndTime := TaskStartTime.Add(JobExcuteTime)
	endTime := TaskEndTime.Format("2006-01-02 15:04:05")
	chtime := changeSubmitTime("2017-12-10 00:00:00",sim.timePer)
	//fmt.Println(chtime)
	if cmp(startTime,"2017-10-03 00:00:00") < 0{
		startTime = "2017-10-03 00:00:00"
	}
	if cmp(startTime,chtime) > 0{
		startTime = chtime
	}
	
	if cmp(endTime,"2017-10-03 00:00:00") < 0{
		endTime = "2017-10-03 00:00:00"
	}
	if cmp(endTime,chtime) > 0{
		endTime = chtime
	}
	excuteTime :=cmp(endTime,startTime)

	sim.mulog.Write([]byte(fmt.Sprintf("%d %d %d\n",rid,gpunum,excuteTime)))
	fmt.Sprintf("%d %d %d\n",rid,gpunum,excuteTime)
}

func (sim *FlowSimulator)FlowResLog( str string){
	sim.reslog.Write([]byte(str))
}

func (sim *FlowSimulator)FlowQueLenLog(curtime string,que *list.List){
	len := que.Len()
	str := ""
	//chtime := changeSubmitTime("2017-12-10 00:00:00",sim.timePer)
	for e:=que.Front();e!=nil;e=e.Next(){
		jinfo,_ :=  (e.Value).(*JobInfo)
		if cmp(jinfo.Submitted_time,"2017-10-03 00:00:00") < 0{
			len--
			continue 
		}
		if cmp(jinfo.Submitted_time,"2017-12-10 00:00:00") > 0{
			len--
			continue
		} 
		if cmp(jinfo.Attempts[0].Start_time,"2017-12-10 00:00:00") > 0{
			len--
			continue
		}
		sec := getTimePast(curtime,jinfo.Submitted_time)
		//if sec > 500000 {
		//	vlog.Dlog(curtime)
		//	vlog.Dlog(fmt.Sprintln(jinfo))
		//}
		str+=fmt.Sprintf(" %d",sec)
	}
	str += "\n"
	sim.qlenlog.Write([]byte(fmt.Sprintf("%d",len)+str))

}

func (sim *FlowSimulator)Init(){
	sim.parser = NewJobParser()
	sim.sh = &SimHeap{}
	heap.Init(sim.sh)
	sim.jtf = new(JobTransFormer)
	sim.jtf.Init()
	sim.jinfos = make([]*JobInfo,0,0)
	sim.desToInfo = make(map[*flowscheduler.JobDescriptor]*JobInfo)
	sim.infoToDes = make(map[*JobInfo]*flowscheduler.JobDescriptor)
	sim.logfile,_ = os.OpenFile("./simulator/FlowSimJob.log",os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	//fd,_:=os.OpenFile("a.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	//sim.logfile.Write([]byte("test"))
	sim.wtlog,_ = os.OpenFile("./simulator/FlowWaitingTime.log",os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	sim.mulog,_ = os.OpenFile("./simulator/FlowMachineUti.log",os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	sim.reslog,_ = os.OpenFile("./simulator/FlowRes.log",os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	sim.qlenlog,_ = os.OpenFile("./simulator/FlowQueLen.log",os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	sim.jobNum = 120000
	sim.rg = new(ResTransformer)
	sim.fs = flowscheduler.NewFlowScheduler()
	sim.queue = list.New()
	sim.timePer = 1.00
}


func (sim *FlowSimulator)GetLogTimes(){
	sTime := "2017-10-03 00:00:00"
	//eTime := "2017-12-15 18:42:00"
	startTime,_ := time.ParseInLocation("2006-01-02 15:04:05",sTime,time.Local)
	//endTime,_ := time.ParseInLocation("2006-01-02 15:04:05",eTime,time.Local)
	for i:=0;i<97920;i++{
		str := fmt.Sprintf(startTime.Format("2006-01-02 15:04:05"))
		si := new(SimInfo)
		si.curTime = str
		si.kind = LOGTIME
		si.jinfo = nil
		heap.Push(sim.sh,si)
		startTime = startTime.Add(time.Minute)
	}
}

func changeSubmitTime(curTime string,fac float64)string{
	stime,err := time.ParseInLocation("2006-01-02 15:04:05","2017-10-03 00:00:00",time.Local)
	if err != nil {
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}
	etime,err := time.ParseInLocation("2006-01-02 15:04:05","2017-12-10 00:00:00",time.Local)
	if err != nil{
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}

	ctime,err := time.ParseInLocation("2006-01-02 15:04:05",curTime,time.Local)
	if err != nil{
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}
	dur := ctime.Sub(stime)
	if dur < 0 &&  ctime.Sub(etime) > 0 {
		return curTime
	}
	tmp := float64(int64(dur))*fac
	fdur := time.Duration(int64(tmp))
	submitTime := stime.Add(fdur)
	ans := submitTime.Format("2006-01-02 15:04:05")
	return ans
}

func (sim *FlowSimulator)SetTimePer(p float64){
	sim.timePer = p
}
func (sim *FlowSimulator)BeforStart(){
	vlog.Vlog(fmt.Sprintf("before start the simulator"))
	go sim.parser.NextJobInfo()
	sim.parser.ch <- sim.jobNum
	<-sim.parser.ch
	sim.jinfos = sim.parser.Jobs
	firstTime := "2017-10-03 00:00:00"
	lastTime := "2017-12-10 00:00:00"
	lastTime = changeSubmitTime("2017-12-10 00:00:00",sim.timePer)
	for _,job := range sim.jinfos {
		per := 1.0
		if cmp(job.Submitted_time,"2017-10-03 00:00:00") >0 && cmp(job.Submitted_time,"2017-12-10 00:00:00") < 0{
			per = sim.timePer
		}
		submitTime := changeSubmitTime( job.Submitted_time,per)
		job.Submitted_time = submitTime
		si := new(SimInfo)
		si.curTime = submitTime
		si.kind = SUBMITTED
		si.jinfo = job
		heap.Push(sim.sh,si)
	}

	heap.Init(sim.sh)
	//sim.initLogTime(firstTime,lastTime)
	dur := cmp(lastTime,firstTime)
	sim.mulog.Write([]byte(fmt.Sprintf("%d\n",dur)))
	ress := RestfMain()
	
	sim.mulog.Write([]byte(fmt.Sprintf("%d\n",len(ress))))
	for i,_ := range ress{
		sim.fs.AddResource(ress[i])
		sim.ResTotal.Gpu += ress[i].ResAvailable.Gpu
		sim.mulog.Write([]byte(fmt.Sprintf("%d %d\n",ress[i].Rid,ress[i].ResAvailable.Gpu)))
	}
	sim.ResAvailable.Gpu = sim.ResTotal.Gpu
	fmt.Printf("all res: %d %d\n",sim.ResTotal.Gpu,sim.ResAvailable.Gpu)
}

func (sim *FlowSimulator)StartSimulate(){
	var i int = 0
	for sim.sh.Len()>0{
		i++
		if i%1000 == 0{
			fmt.Println(i)
		} 
		//sim.logfile.Write([]byte(fmt.Sprintf("%d ",i)))
		si := heap.Pop(sim.sh).(*SimInfo)
		jinfo := si.jinfo
		var solverRunTime time.Duration
		if si.kind == COMPLETED {
			jd := sim.infoToDes[jinfo]
			sim.HandleJobCompleted(jd)
			
			taskToRes := make(map[flowscheduler.TaskID]flowscheduler.ResID)
			
			if sim.ResAvailable.Gpu != 0{
				startTime := time.Now()
				taskToRes = sim.fs.ScheduleAllJobs()
				endTime := time.Now()
				solverRunTime = endTime.Sub(startTime)
			}
			sim.HandleScheduled(si.curTime ,taskToRes,solverRunTime)
			sim.LogInfo(si,taskToRes)
		} else if si.kind == SUBMITTED{
			jd := sim.jtf.Transform(jinfo)
			sim.infoToDes[jinfo] = jd
			sim.desToInfo[jd] = jinfo
			sim.fs.AddJob(jd)
			sim.queue.PushBack(jinfo)
			taskToRes := make(map[flowscheduler.TaskID]flowscheduler.ResID)
			if sim.ResAvailable.Gpu != 0{
				startTime := time.Now()
				taskToRes = sim.fs.ScheduleAllJobs()
				endTime := time.Now()
				solverRunTime = endTime.Sub(startTime)
			}
			//fmt.Println(taskToRes)
			sim.HandleScheduled(si.curTime,taskToRes,solverRunTime)
			sim.LogInfo(si,taskToRes)
		} else if si.kind == LOGTIME {
			sim.FlowResLog(fmt.Sprintf("%d\n",sim.ResTotal.Gpu - sim.ResAvailable.Gpu))
			sim.FlowQueLenLog(si.curTime, sim.queue)
		}
	}
}

func (sim *FlowSimulator)HandleScheduled(curTime string,taskToRes map[flowscheduler.TaskID]flowscheduler.ResID,sloverRunTime time.Duration){
	for tid,_ := range taskToRes {
		jd := sim.fs.TaskIDToJobDes(tid)
		jinfo := sim.desToInfo[jd]

		for e:=sim.queue.Front();e!=nil;e=e.Next(){
			if e.Value == jinfo{
				sim.queue.Remove(e)
			}
		}

		excuteTime := jinfo.ExcuteTime()
		startTime,err := time.ParseInLocation("2006-01-02 15:04:05",curTime,time.Local)
		if err != nil {
			vlog.Vlog("Error, sim.HandleScheduled")
			fmt.Println(err)
		}
		
		
		//sim.FlowWaitingTimeLog(curTime,jinfo)
		//sim.FlowMachineUtiLog(curTime,jinfo)

		gpunum := jinfo.GetGpuNeedNum()
		taskEndTime := startTime.Add(excuteTime)
		endTime := taskEndTime.Format("2006-01-02 15:04:05")
		siminfo := &SimInfo{endTime,jinfo,COMPLETED}
		//fmt.Println(startTime,endTime,excuteTime)
		heap.Push(sim.sh,siminfo)
		sim.ResAvailable.Gpu -= gpunum
	}
}
func (sim *FlowSimulator)LogInfo(si *SimInfo,taskToRes map[flowscheduler.TaskID]flowscheduler.ResID){
	str := ""
	str += si.curTime
	gpu := fmt.Sprintf("%d ",si.jinfo.GetGpuNeedNum())
	if si.kind == SUBMITTED {
		str += " " + si.jinfo.Jobid + " " + "Submitted"+"(gpu need: "+gpu+ ")"
	} else if si.kind == COMPLETED{
		str += " " + si.jinfo.Jobid + " " + "Completed"+"(gpu need: "+gpu+ ")"
	} else {
		str += "error ,simInfo.kind is illeagl"
		sim.logfile.Write([]byte(str))
		return 
	}
	qi,_ := sim.GetWaittingInfo()
	str += qi

	
	for tid,rid := range taskToRes{
		jd := sim.fs.TaskIDToJobDes(tid)
		rd := sim.fs.ResIDToResDes(rid)
		gpunum := jd.Tasks[0].ResRequest.Gpu
		str += fmt.Sprintf("task: %d(gpu need: %d) is scheduled to res :%d(gpu avai: %d)\n",tid,gpunum,rid,rd.ResAvailable.Gpu)
		jinfo,_ := sim.desToInfo[jd]
		//att1 := jinfo.Attempts[0]
		//att2 := jinfo.Attempts[len(jinfo.Attempts)-1]
		sim.FlowMachineUtiLog(rid,si.curTime,jinfo)
		sim.FlowWaitingTimeLog(si.curTime,jinfo)
		//excuteTime:=cmp(endTime,startTime)
		//sim.mulog.Write([]byte(fmt.Sprintf("%d %d %d\n",rd.Rid,gpunum,excuteTime)))
	}
	
	sim.logfile.Write([]byte(str))
	//sim.logfile.Write([]byte(fmt.Sprintln(taskToRes)))
	return 
}

//"2017-12-15"
func (sim *FlowSimulator)GetWaittingInfo()(string,int){
	ans := "details: "
	var sum uint64  = 0
	len := 0 
	for e:= sim.queue.Front();e != nil;e = e.Next() {
		jinfo := e.Value.(*JobInfo)
		jd,ok := sim.infoToDes[jinfo]
		if !ok {
			vlog.Dlog("error ,there are no jobdes for jinfo")
			return "",0
		}
		for _,td := range jd.Tasks{
			if td.State == flowscheduler.TASK_UNSCHEDULED{
				sum += td.ResRequest.Gpu
				ans += fmt.Sprintf("%d ",td.ResRequest.Gpu)
				len++
			}
		}
	}
	return fmt.Sprintf(" wq len:%d, all Gpu need: %d, ResAvailable :%d ",len,sum,sim.ResAvailable.Gpu)+ans+"\n",len
}

func (sim *FlowSimulator)HandleJobCompleted(jd *flowscheduler.JobDescriptor){
	sim.fs.HandleJobCompleted(jd)
	jinfo := sim.desToInfo[jd]
	sim.ResAvailable.Gpu += jinfo.GetGpuNeedNum()
	delete(sim.desToInfo,jd)
	delete(sim.infoToDes,jinfo)
}
func FlowSimuMain(){
	fmt.Sprintf("This is the flow simulator main!\n")
	fmt.Println("FlowSimMain()")
	sim := new(FlowSimulator)
	sim.Init()
	sim.GetLogTimes()
	sim.BeforStart()
	fmt.Println(sim.ResAvailable,sim.ResTotal)
	//fmt.Println("QAQ")
	//fmt.Println(sim.sh)
	sim.StartSimulate()
	//fmt.Println(sim.desToInfo)
	//fmt.Println(sim.infoToDes)
}




