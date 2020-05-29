package simulator

import (
	"fmt"
	"time"
	"sort"
	"vlog"
	"os"
	"container/list"
)

func getTimePast(isTime,jsTime string)int64{
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
	


type timePoints []string

func (tp  timePoints)Len()int{return len(tp)}
func (tp  timePoints)Swap(i,j int){tp[i],tp[j] = tp[j],tp[i]}
func (tp  timePoints)Less(i,j int)bool{
	itime,err := time.ParseInLocation("2006-01-02 15:04:05",tp[i],time.Local)
	if err != nil {
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}
	jtime,err := time.ParseInLocation("2006-01-02 15:04:05",tp[j],time.Local)
	if err != nil{
		vlog.Vlog("Error, tp.less")
		fmt.Println(err)
	}

	ans := itime.Sub(jtime)
	return ans < 0 
}

type JobSim struct {
	parser *JobParser
	tp  timePoints
	timeToJob map[string][]*JobInfo
	jinfo  []*JobInfo
	joblog *os.File
	wtlog *os.File
	reslog,queuelog,mutilog *os.File
	jobNum int
	gputotal,gpuavi uint64//GpuTotal,GpuAvailable uint64
	//2348 total 
	logtimes map[string]bool
	gpus map[string]int
}

func (js *JobSim)Init(){
	js.parser = NewJobParser()
	js.tp = timePoints{}
	js.timeToJob = make(map[string][]*JobInfo)
	js.jinfo = make([]*JobInfo,0,0)
	
	js.joblog,_ = os.OpenFile("./simulator/JobLog.log",os.O_RDWR|os.O_CREATE,0644)
	js.wtlog,_ = os.OpenFile("./simulator/WaitingTime.log",os.O_RDWR|os.O_CREATE,0644)
	js.reslog,_ = os.OpenFile("./simulator/Res.log",os.O_RDWR|os.O_CREATE,0644)
	js.queuelog,_ = os.OpenFile("./simulator/QueLen.log",os.O_RDWR|os.O_CREATE,0644)
	js.mutilog,_ = os.OpenFile("./simulator/MachineUti.log",os.O_RDWR|os.O_CREATE,0644)

	//fd,_:=os.OpenFile("a.txt",os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	js.jobNum = 120000
	js.logtimes = make(map[string]bool)
	js.gputotal = 2364
	js.gpuavi = 2364
	js.gpus = make(map[string]int)
}



func (js *JobSim)JobWaitingTime(jinfo *JobInfo)float64{
	att := jinfo.Attempts[0]
	//att := jif.Attempts[len(jif.Attempts)-1]
	startTime,err := time.ParseInLocation("2006-01-02 15:04:05",att.Start_time,time.Local)
	if err != nil {
		vlog.Vlog("Error, JobSim.JobWaitingTime")
		fmt.Println(err)
	}
/*
	endTime,err := time.ParseInLocation("2006-01-02 15:04:05",att.End_time,time.Local)
	if err != nil{
		vlog.Vlog("Error, JobSim.JobWaitingTime")
		fmt.Println(err)
	}
*/
	submitTime,err := time.ParseInLocation("2006-01-02 15:04:05",jinfo.Submitted_time,time.Local)
	if err != nil{
		vlog.Vlog("Error, JobSim.JobWaitingTime")
		fmt.Println(err)
	}
	//gpuneed := jinfo.GetGpuNeedNum()

	dur := startTime.Sub(submitTime)
	//wt := fmt.Sprintf("%d %.0f\n",gpuneed,dur.Seconds())
	//js.wtlog.Write([]byte(wt))
	return dur.Seconds()
}

func (js *JobSim)GetLogTimes(){
	sTime := "2017-10-03 00:00:00"
	//eTime := "2017-12-15 18:42:00"
	startTime,_ := time.ParseInLocation("2006-01-02 15:04:05",sTime,time.Local)
	//endTime,_ := time.ParseInLocation("2006-01-02 15:04:05",eTime,time.Local)
	for i:=0;i<97920;i++{
		str := fmt.Sprintf(startTime.Format("2006-01-02 15:04:05"))
		js.logtimes[str] = true
		js.tp =append(js.tp,str)
		startTime = startTime.Add(time.Minute)
	}
}

func (js *JobSim)WaitingTimeLog(job *JobInfo){
	if cmp(job.Submitted_time,"2017-10-03 00:00:00") < 0{
		return 
	}
	if cmp(job.Submitted_time,"2017-12-10 00:00:00") > 0{
		return
	} 
	if cmp(job.Attempts[0].Start_time,"2017-12-10 00:00:00") > 0{
		return
	}
	wtime := js.JobWaitingTime(job)
	js.wtlog.Write([]byte(fmt.Sprintf("%d %.0f\n",job.GetGpuNeedNum(),wtime)))
}

func (js *JobSim)ResLog(str string){
	js.reslog.Write([]byte(str))
}

func (js *JobSim )QueueLog(curtime string, que *list.List){
	len := que.Len()
	str := ""
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
	js.queuelog.Write([]byte(fmt.Sprintf("%d",len)+str))
}


func (js *JobSim) MutiLog(job *JobInfo){
	startTime := job.Attempts[0].Start_time
	endTime := job.Attempts[len(job.Attempts)-1].End_time
	if cmp(startTime,"2017-10-03 00:00:00") < 0{
		startTime = "2017-10-03 00:00:00"
	}
	if cmp(startTime,"2017-12-10 00:00:00") > 0{
		startTime = "2017-12-10 00:00:00"
	}
	
	if cmp(endTime,"2017-10-03 00:00:00") < 0{
		endTime = "2017-10-03 00:00:00"
	}
	if cmp("2017-12-10 00:00:00",endTime) < 0{
		endTime = "2017-12-10 00:00:00"
	}
	excuteTime:=cmp(endTime,startTime)
	gpunum  := job.GetGpuNeedNum()
	js.mutilog.Write([]byte(fmt.Sprintf("%d %d\n",gpunum,excuteTime)))
	fmt.Sprintf("%d %.0f\n",gpunum,excuteTime)

}
func (js *JobSim)BeforeStart(){
	vlog.Vlog(fmt.Sprintf("before start the simulator"))
	go js.parser.NextJobInfo()
	sum := 0
	js.GetLogTimes()
	js.parser.ch <- js.jobNum
	<-js.parser.ch
	js.jinfo = js.parser.Jobs
	for _,job := range js.jinfo {
		
		att1 := job.Attempts[0]
		att2 := job.Attempts[len(job.Attempts)-1]
		startTime := att1.Start_time
		endTime  := att2.End_time
		submitTime := job.Submitted_time
		
		if _,ok := js.timeToJob[startTime];!ok{
			js.timeToJob[startTime] = make([]*JobInfo,0,0)
			js.tp = append(js.tp,startTime)
		}
		js.timeToJob[startTime] = append(js.timeToJob[startTime],job)

		if _,ok := js.timeToJob[submitTime];!ok{
			js.timeToJob[submitTime] = make([]*JobInfo,0,0)
			js.tp = append(js.tp,submitTime)
		}
		js.timeToJob[submitTime] = append(js.timeToJob[submitTime],job)

		if _,ok := js.timeToJob[endTime];!ok{
			js.timeToJob[endTime] = make([]*JobInfo,0,0)
			js.tp = append(js.tp,endTime)
		}
		js.timeToJob[endTime] = append(js.timeToJob[endTime],job)
	}
	sort.Sort(js.tp)
	//fmt.Println(len(js.gpus))
	vlog.Vlog(fmt.Sprintf("end sim.BeforStart,there are %d has question",sum))
	fmt.Sprintf("end sim.BeforStart,there are %d has question\n",sum)
}
func (js *JobSim)JobLog(str string, wq *list.List){
	str += fmt.Sprintf("gpuAvai: %d, ",js.gputotal - uint64(len(js.gpus)))
	totalneed, detail := GpuNeedInQueue(wq)
	str+= totalneed
	str += detail
	//str += GenerateQueueDetail(waitqueue)
	js.joblog.Write([]byte(str))
}

func (js *JobSim)StartSim(){
	vlog.Vlog(fmt.Sprintf("Start the Simulation"))
	fmt.Println("start sim!")
	usedNum := 0 
	js.joblog.Write([]byte("start the job sim,wq(waiting queue),gpu(all gpu need in waiting queue)\n"))
	waitqueue := list.New()
	lastTime := "2016-01-02 03:04:05"
	for _,curtime := range js.tp {
		//vlog.Dlog(curtime)
		if lastTime == curtime{
			continue
		} else {
			lastTime = curtime
		}
		for _,jinfo := range js.timeToJob[curtime]{
			var str string	
			if jinfo.Submitted_time == curtime{
				waitqueue.PushBack(jinfo)
				str = fmt.Sprintf(curtime+" "+jinfo.Jobid+" Subbmit, wq len:%d,",waitqueue.Len())
			} else {
				att1 := jinfo.Attempts[0]
				att2 := jinfo.Attempts[len(jinfo.Attempts)-1]
				if att1.Start_time == curtime{
					for e:=waitqueue.Front();e!=nil;e=e.Next(){
						if e.Value == jinfo{
							waitqueue.Remove(e)
							break
						}
					}
					//js.gpuavi -= jinfo.GetGpuNeedNum()
					usedNum += int(jinfo.GetGpuNeedNum())
					tmp := jinfo.GetGpuDetail()
					for _,v := range tmp{
						if _,ok := js.gpus[v];!ok{
							js.gpus[v] = 1
						} else {
							js.gpus[v] += 1
						}
					}
					js.MutiLog(jinfo)
					js.WaitingTimeLog(jinfo)

					str = fmt.Sprintf(curtime+" "+jinfo.Jobid+" Scheduled, wq len:%d,",waitqueue.Len())
				} else if att2.End_time == curtime{
					//js.gpuavi += jinfo.GetGpuNeedNum()
					tmp := jinfo.GetGpuDetail()
					for _,v := range tmp{
						if _,ok := js.gpus[v];!ok{
							vlog.Dlog("error ! js.start simulator \n,this gpu should be there\n")
						} else {
							js.gpus[v] -= 1
						}
						if val,_ := js.gpus[v];val<=0{
							delete(js.gpus,v)
						}
					}
					//js.gpuavi += jinfo.GetGpuNeedNum()
					usedNum -= int(jinfo.GetGpuNeedNum())
					str = fmt.Sprintf(curtime+" "+jinfo.Jobid+" Completed, wq len:%d,",waitqueue.Len())
				}
			}
			js.JobLog(str,waitqueue)
		}
		if _,ok := js.logtimes[curtime];ok{
			//detail := fmt.Sprintf("%d %d\n",waitqueue.Len(),len(js.gpus))
			if usedNum > 2364 {
				usedNum = 2364
			}
			js.ResLog(fmt.Sprintf("%d\n",usedNum))
			js.QueueLog(curtime,waitqueue)
			delete(js.logtimes,curtime)
		}
		delete(js.timeToJob,curtime)
	}
}

func GpuNeedInQueue(l *list.List)(string,string){
	var ans uint64 = 0
	detail := "details : "
	for e:=l.Front();e!=nil;e= e.Next(){
		jif := e.Value.(*JobInfo)
		num := jif.GetGpuNeedNum()
		ans += num
		detail += fmt.Sprintf("%d ",num)
	}
	detail+="\n"
	return fmt.Sprintf("wq total: %d. ",ans),detail
}
func GenerateQueueDetail(wq *list.List)string{
	
	return ""
}

func SimJobMain(){
	js := new(JobSim)
	js.Init()
	fmt.Println("SimJobMain() start!")
	js.BeforeStart()
	//fmt.Println(len(js.tp))
	//fmt.Println("------------------")
	js.StartSim()
	
}

