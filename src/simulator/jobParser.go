package simulator

//import "flowgraph"
import (
	"fmt"
	"vlog"
	"encoding/json"
	"os"
	"time"
	"strconv"
	//"reflect"
	//"sort"
)
/*type strarr []string  
func (s strarr)Len()int{
	return len(s)
}
func (s strarr)Swap(i,j int){
	s[i],s[j] = s[j],s[i]
}
func (s strarr)Less(i,j int)bool{
	return s[i]<s[j]
}*/
type JobParser struct {
	// 通过ch接收需要接下来的num个jobinfo，将max（num，剩余的job个数）的信息储存在Jobs
	ch chan int
	Jobs []*JobInfo
}

func NewJobParser()*JobParser{
	sim := new(JobParser)
	sim.Jobs = make([]*JobInfo,0,0)
	sim.ch = make(chan int)
	return sim
}

type DetailInfo struct {
    Ip string `json:"ip"`
    Gpus []string  `json:"gpus"`
}

type AttemptInfo struct{
    Start_time string `json:"start_time"`
    End_time string `json:"end_time"`
    Detail []DetailInfo `json:"detail"`
}

type JobInfo struct{
    Status string `json:"status"`
    Vc string `json:"vc"`
    Jobid string `json:"jobid"`
    Attempts []AttemptInfo   `json:"attempts"`
    Submitted_time string `json:"submitted_time"`
	User string `json:"user"`
	Test string `json:"server"`
}

func checkDetail(att AttemptInfo,tem map[string]bool)bool{
	for _,v := range att.Detail{
		ip := v.Ip
		if ip == ""{
			return true
		}
		tem[ip] = true
		num,_ := strconv.Atoi(ip[1:])
		if num >= 424{
			fmt.Println(ip)
			return false
		}
	}
	return true
}
func (jif *JobInfo)ExcuteTime()time.Duration{
	att1 := jif.Attempts[0]
	att2 := jif.Attempts[len(jif.Attempts)-1]
	startTime,err := time.ParseInLocation("2006-01-02 15:04:05",att1.Start_time,time.Local)
	if err != nil {
		vlog.Vlog("Error, jobinfo.Excutetime")
		fmt.Println(err)
	}

	endTime,err := time.ParseInLocation("2006-01-02 15:04:05",att2.End_time,time.Local)
	if err != nil{
		vlog.Vlog("Error, jobinfo.Excutetime")
		fmt.Println(err)
	}

	ans := endTime.Sub(startTime)
	//fmt.Println(ans.Seconds())
	//fmt.Println(startTime,endTime,startTime.Add(ans))
	return ans
}

func (jif *JobInfo)GetGpuNeedNum()uint64{
	var ans uint64 = 0
	for _,det := range jif.Attempts[len(jif.Attempts)-1].Detail{
		ans+=uint64(len(det.Gpus))
	}
	return ans
}

func (jif *JobInfo)GetGpuDetail()[]string{
	ans := make([]string,0,0)
	//for _,att := range jif.Attempts{
		att := jif.Attempts[len(jif.Attempts)-1]
		for _,det := range att.Detail{
			for _,gn := range det.Gpus{
				ans = append(ans,det.Ip+gn)
			}
		}
	//}
	return ans
}

func (sim *JobParser)NextJobInfo(){
	fmt.Sprintf("test of flow graph")
	vlog.Vlog("simulattor start")
	file,err := os.Open("simulator/cluster_job_log")
	defer file.Close()
	if err != nil {
		fmt.Println("can not open file\n")
	} 
	dec := json.NewDecoder(file)
	t,err := dec.Token()
	if err != nil {
		fmt.Println(err)
	}
	//tem := make(map[string]bool)
	for {
		num := <- sim.ch
		sim.Jobs = make([]*JobInfo,0,0)
		if num == -1 || num == 0 || !dec.More(){
			break
		}
		//fmt.Println(num)
		for ;num>0 && dec.More();num--{
			jinfo := new(JobInfo)
			if err = dec.Decode(jinfo); err != nil {
				fmt.Println(err)
				fmt.Println("error\n");
			}
			if len(jinfo.Attempts) == 0{
				continue
			}
			att1 := jinfo.Attempts[0]
			att2 := jinfo.Attempts[len(jinfo.Attempts)-1]
			startTime := att1.Start_time
			endTime  := att2.End_time
			submitTime := jinfo.Submitted_time
			if startTime == "None" || endTime == "None" || submitTime == "None" {
				continue
			}
			gpunum := jinfo.GetGpuNeedNum()
			if gpunum == 0 { 
				continue
			}
			sim.Jobs = append(sim.Jobs,jinfo)
			//fmt.Println(jinfo)	
			//checkDetail(att1,tem)
			//checkDetail(att2,tem)

		}
		/*ms := strarr{}
		for k,_ := range tem{
			ms = append(ms,k)
		}
		fmt.Println(ms)
		sort.Sort(ms)
		for _,v := range ms {
			vlog.Dlog(v)
		}*/
		sim.ch <- len(sim.Jobs)
	}	
	

	t,err = dec.Token()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Sprintf("%T: %v\n", t, t)
	vlog.Vlog("exit job transformer")
	return 
}

func (jp *JobParser)SetNum(n int){
	jp.ch<-n
}

func (jp *JobParser)GetJobs()[]*JobInfo{
	<-jp.ch
	return jp.Jobs
}