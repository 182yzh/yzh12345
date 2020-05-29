package simulator

import "flowscheduler"
import "os"
import "strconv"
import "fmt"
import "strings"
import "bufio"
import "io"

type ResTransformer struct{
	curid flowscheduler.ResID
	unusedid map[flowscheduler.ResID]bool
} 

func (rtf *ResTransformer)NextID()flowscheduler.ResID{
	var next flowscheduler.ResID
	if len(rtf.unusedid ) == 0{
		rtf.curid++
		next = rtf.curid
	} else {
		for id,_ := range rtf.unusedid{
			next = id
			break
		}
	}
	return next
}


func (rtf *ResTransformer)GenerateResDes(n uint64)*flowscheduler.ResDescriptor{
	rd := new(flowscheduler.ResDescriptor)
	rv := &flowscheduler.ResVector{100,100,n}
	rd.ResAvailable = *rv
	rd.ResTotal = *rv
	rd.CurrentRunningTasks = make([]flowscheduler.TaskID,0,0)
	rd.Rid = rtf.NextID()
	rd.Rtype = flowscheduler.RES_PU
	return rd
}


func (rtf *ResTransformer)ReadFile()[]*flowscheduler.ResDescriptor{
	file,err := os.Open("./simulator/cluster_machine_list")
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return nil
    }
    defer file.Close()

	br := bufio.NewReader(file)
	br.ReadLine()
	ress := make([]*flowscheduler.ResDescriptor,0,0)
    for {
        buff,_,c := br.ReadLine()
        if c == io.EOF {
            break
        }
		info := string(buff)
		details := strings.Split(info,",")
		num, err := strconv.ParseInt(details[1], 10, 64)
		if err != nil{
			fmt.Printf("%s\n",err)
		}
		rd := new(flowscheduler.ResDescriptor)
		rd.Rid = rtf.NextID()
		rd.Rtype = flowscheduler.RES_PU
		rd.ResAvailable.Gpu = uint64(num)
		rd.ResTotal.Gpu = uint64(num)
		ress = append(ress,rd)
	}
	return ress
}

func RestfMain()[]*flowscheduler.ResDescriptor{
	rtf := new(ResTransformer)
	rtf.unusedid = make(map[uint64]bool)
	ans := rtf.ReadFile()
	return ans
}