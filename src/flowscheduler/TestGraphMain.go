package flowscheduler

import (
	"fmt"
	"time"

)
var curid uint64
func GenerateRes(resavi map[uint64]uint64)[]*ResDescriptor{
	ans := make([]*ResDescriptor,0,0)
	var i uint64 = 0
	for k,v := range resavi{
		for i =0 ;i<v;i++{
			ans = append(ans,NewRes(k))
		}
	}
	return ans
}

func GenerateJob(resreq map[uint64]uint64)[]*JobDescriptor{
	ans := make([]*JobDescriptor,0,0)
	var i uint64 
	for k,v := range resreq{
		for i =0;i<v;i++{
			ans = append(ans,NewJob(k))
		}
	}
	return ans
}

func NewJob(req uint64 )*JobDescriptor{
	jd := new(JobDescriptor)
	td := new(TaskDescriptor)
	jd.Jid = curid 
	curid ++ 
	td.Tid = curid 
	curid ++ 
	jd.Tasks = make([]*TaskDescriptor,0,0)
	jd.Tasks = append(jd.Tasks,td)
	jd.Name = fmt.Sprintf("job-%d",jd.Jid)
	td.Name = fmt.Sprintf("job-%d task-%d",jd.Jid,td.Tid)
	jd.State = JOB_CREATED
	td.State = TASK_UNSCHEDULED
	td.Jd = jd
	td.SubmitTime = uint64(time.Now().UnixNano())
	jd.SubmitTime = uint64(time.Now().UnixNano())
	td.ResRequest.Gpu = req
	fmt.Println(fmt.Sprintf("task:%d,gpu need:%d",td.Tid,req))
	return jd
}

func NewRes(avi uint64)*ResDescriptor{
	rd := new(ResDescriptor)
	rd.ResAvailable.Gpu = avi
	rd.Rid = curid
	curid ++ 
	rd.ResTotal.Gpu = avi
	rd.Rtype = RES_PU
	fmt.Println(fmt.Sprintf("res:%d, gpu avi:%d",rd.Rid,avi))
	return rd
}


func TestGraphMain(){
	//return
	curid = 1
	resreq := make(map[uint64]uint64)
	resreq[16]=2
	resreq[1]=10

	resavi := make(map[uint64]uint64)
	resavi[2]=150
	resavi[4]=0
	resavi[8]=50
	ress := GenerateRes(resavi)
	jobs := GenerateJob(resreq)

	fs := NewFlowScheduler()
	for _,v := range ress{
		fs.AddResource(v)
	}
	fs.AddJobs(jobs)
	shc := fs.ScheduleAllJobs()
	fmt.Println(shc)
	fmt.Println("num: ",len(shc))
}