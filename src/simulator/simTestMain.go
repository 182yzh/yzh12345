package simulator

import "fmt"
import "vlog"
import "time"

func SimTestMain(){
	vlog.Vlog(fmt.Sprintf("start the test of simulator"))
	sim := NewJobParser()
	go sim.NextJobInfo()
	sum := 0 
	earlist := time.Now()
	for {
		sim.ch <- 500
		sum += <-sim.ch
		//fmt.Println(num)
		for _,jobinfo := range sim.Jobs{
			//fmt.Println(jobinfo.Submitted_time)
			submit_time,err := time.ParseInLocation("2006-01-02 15:04:05",jobinfo.Submitted_time,time.Local)
			if err != nil {
				fmt.Println(err)
				return 
			}
			if earlist.Sub(submit_time) > 0 {
				earlist = submit_time
			}
			//,jobinfo.Attempts
		}
		break;
	}

	for {
		sim.ch <- 850
		sum += <-sim.ch
		//fmt.Println(num)
		for i,jobinfo := range sim.Jobs{
			//fmt.Println(jobinfo)
			//fmt.Println(jobinfo.Submitted_time)
			if len(jobinfo.Attempts) == 0{
				fmt.Println(i,jobinfo)
				continue
			}
			att := jobinfo.Attempts[len(jobinfo.Attempts)-1]
			//fmt.Println(att.Start_time,att.End_time)
			fmt.Sprintf(att.End_time)
		}
		break;
	}
	fmt.Println(earlist,"all task num =",sum)
}