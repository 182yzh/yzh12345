package flowscheduler

import "fmt"
import "bufio"
import "strings"
import "os/exec"
import "io/ioutil"
import "container/list"
import "sort"
import "vlog"
import "time"

type Dispatcher struct{
	gm *GraphManager
	runOnce bool
} 


func NewDispatcher(gm *GraphManager)*Dispatcher{
    vlog.Vlog("New Dispatcher")
    return &Dispatcher{
        runOnce:false,
        gm:gm,
    }
}



// 返回solver的输出
func (dp *Dispatcher)RunSolver() (string,map[NodeID]NodeID) {
    vlog.Vlog("dispatcher - Run Solver")
    switch dp.gm.costmodel.(type){
    case *GPUCostModel:
        ch := make(chan map[NodeID]NodeID)
        //go dp.RunPinPackingSlover(ch)
        go dp.RunGpuSlover(ch)
        taskMappings := <-ch 
        return "==+++==",taskMappings
        //return dp.RunGpuSlover()
    default:
        vlog.Vlog("Unknow cost models")
    }
    dp.gm.sinkNode.excess = 0
    dp.gm.UpdateFlowGraph()
    //fmt.Println("update flow graph")
    input := dp.gm.ExportGraph()
    //fmt.Println("__\n"+input+"__\n")
    //vlog.Temlog(input)
    cmd := exec.Command("./flowscheduler/slover/cs2.exe")
    defer cmd.Wait()
    stdout, err := cmd.StdoutPipe()
    defer stdout.Close()
	if err != nil{
        return "",nil
	}
    
    stdin, err := cmd.StdinPipe()
    //defer stdin.Close()
	if err != nil{
        return "",nil
    }
    cmd.Start()
    stdin.Write([]byte(input))
    stdin.Close()
    out_bytes, _ := ioutil.ReadAll(stdout)
    
    //out_bytes,_ := cmd.Output()
    output := string(out_bytes)
    //fmt.Println(output)
    //vlog.Temlog(output)
    flow := dp.ParseOutput(output)
    //fmt.Println(flow)
    taskMappings := dp.GetMappings(flow) 
    //fmt.Println(taskMappings)
    dp.runOnce = true
    return output,taskMappings
}

func (dp *Dispatcher)GetMappings(flow map[NodeID](map[NodeID]uint64))map[NodeID]NodeID{
    taskMappings := make(map[NodeID]NodeID)
    leaves := dp.gm.leafNodes
    //fmt.Println(leaves)
    for puid,_ := range leaves {
        if _,ok := flow[puid];!ok{
            continue
        }
        toVisit := list.New()
        toVisit.PushBack(puid)
        for ;toVisit.Len() >0;{
            first := toVisit.Front()
            dst := first.Value.(NodeID)
            toVisit.Remove(first)
            node := dp.gm.gcm.GetNode(dst)
            if node.IsTaskNode() {
                taskMappings[dst]=puid
            } else {
                for src,_ := range flow[dst]{
                    toVisit.PushBack(src)
                }
            }
        }
    }
    return taskMappings
}


func (dp *Dispatcher)ParseOutput(output string )map[NodeID](map[NodeID]uint64){
    vlog.Vlog("dsipatcher -parse solver output")
    ioreader := strings.NewReader(output)
    reader := bufio.NewReader(ioreader)
    flow := make(map[NodeID](map[NodeID]uint64))
    var str string
    var src,dst,f uint64
    for ;; {
        data,_,err := reader.ReadLine();
        if err == nil {
            str = string(data)
            if str[:1] == "f"{
                fmt.Sscanf(str,"f %d %d %d",&src,&dst,&f)
                //fmt.Println(src,dst,flow)
                if f > 0{
                    //fmt.Sprintf(str)
                    if _,ok:=flow[NodeID(dst)];!ok{
                        flow[NodeID(dst)] = make(map[NodeID]uint64)
                    }
                    flow[NodeID(dst)][NodeID(src)] = f
                }
            }
        } else {
            break
        }
    }
    //fmt.Println(flow)
    return flow
}



type  GpuNums []uint64
func (nums GpuNums)Len()int{return len(nums)}
func (nums GpuNums)Swap(i,j int){nums[i],nums[j] = nums[j],nums[i]}
func (nums GpuNums)Less(i,j int)bool{return nums[i]>nums[j]}

func (dp *Dispatcher)RunGpuSlover(ch chan map[NodeID]NodeID)(string,map[NodeID]NodeID){
    vlog.Vlog("Run GPUCostModel Slover")
    nums := GpuNums{}
    for tnode,_ := range dp.gm.queue{
        td := tnode.td    
        if td.State == TASK_RUNNING {
            continue
        }
        nums = append(nums,td.ResRequest.Gpu)
    }
    sort.Sort(nums)
  
    for _,rnode := range dp.gm.resNodes{
        rnode.rd.ResReserved.Gpu = 0
    }
    var cur uint64 = 1<<63
    taskMappings := make(map[NodeID]NodeID)
    vlog.Temlog("task to scedule:" + fmt.Sprintln(nums))
    for _,num := range nums{
        if cur == num{
            continue
        }
        //vlog.Temlog(fmt.Sprintln("process: ",num))
        cur = num
        gcs :=dp.gm.costmodel.(*GPUCostModel) 
        gcs.SetGpuNum(cur)
        dp.gm.sinkNode.excess = 0
        
        dp.gm.UpdateFlowGraph()
       
        _,temTaskMapping := dp.GPULimitSlover()
        
        for tid,rid :=range temTaskMapping{
            tnode := dp.gm.GetNode(tid)
            if _,ok := dp.gm.taskRuningArcs[tnode.td.GetTaskID()];ok {
                continue
            }
            taskMappings[tid]=rid
            rnode := dp.gm.GetNode(rid)
            rnode.rd.ResAvailable.Gpu -= cur
            rnode.rd.ResReserved.Gpu += cur
        }
       
    }
    for _,rnode := range dp.gm.resNodes{
        rnode.rd.ResAvailable.Gpu += rnode.rd.ResReserved.Gpu 
        rnode.rd.ResReserved.Gpu = 0
    }
    ch<-taskMappings
    return "",taskMappings;
}


func (dp *Dispatcher)GPULimitSlover()(string,map[NodeID]NodeID){
    //taskMappings := make(map[NodeID]NodeID)
    input := dp.gm.ExportGraph()
    vlog.Temlog(fmt.Sprintln("__\n"+input+"__\n"))
    
    cmd := exec.Command("./flowscheduler/slover/cs2.exe")
    defer cmd.Wait()
    stdout, err := cmd.StdoutPipe()
	if err != nil{
        fmt.Println(err)
        return "",nil
    }
    defer stdout.Close()
    
    stdin, err := cmd.StdinPipe()
	if err != nil {
        fmt.Println(err)
        return "",nil
    }
    startTime := time.Now()
    cmd.Start()
    stdin.Write([]byte(input))
    stdin.Close()
    out_bytes, err := ioutil.ReadAll(stdout)
    endTime := time.Now()
    endTime.Sub(startTime)
    //vlog.Dlog(fmt.Sprintf("%d",dur.Nanoseconds()/1000000))
    if err != nil {
        fmt.Println(err)
        vlog.Dlog(fmt.Sprintln(err))
        vlog.Dlog("input\n")
        vlog.Dlog(input)
        vlog.Dlog("output\n")
        vlog.Dlog(string(out_bytes))
        vlog.Dlog("--------------------------------\n")
        return "",nil
    }
    //out_bytes,_ := cmd.Output()
    output := string(out_bytes)
    temstrs := strings.Split(output,"\n")
    temoutput := ""
    if len(temstrs) <= 16 {
        fmt.Println(output)
        vlog.Dlog("len temstrs < 16\n")
        vlog.Dlog("input\n")
        vlog.Dlog(input)
        vlog.Dlog("output\n")
        vlog.Dlog(output)
        vlog.Dlog("---------------\n")
        return "",nil
    }
    for _,v := range temstrs[16:]{
        temoutput+= v + "\n"
    }
    vlog.Temlog(fmt.Sprintln("\npart output:\n",temoutput))
    flow := dp.ParseOutput(output)
    taskMappings := dp.GetMappings(flow) 
    vlog.Temlog("taskmappings "+ fmt.Sprintln(taskMappings))
    dp.runOnce = true
    return "",taskMappings
}


func (dp *Dispatcher)ExportGraphWithoutRunningTasks()string{
    return dp.gm.ExportGraphWithoutRunningTasks()
}


















/*

func (dp *Dispatcher) RunPinPackingSlover(ch chan map[NodeID]NodeID )map[NodeID]NodeID{
    woods := make([]PinPacking,0,0)
    boxs := make([]PinPacking,0,0)
    for _,tnode := range dp.gm.taskToNodeMap{
        td := tnode.taskDes
        if td.taskType == TASK_ROOT{
            continue
        }
        woods = append(woods,PinPacking{td.resourceRequest.gpu,tnode.id})
    }

    for _,rnode := range dp.gm.resToNodeMap {
        rd := rnode.resDes
        boxs = append(boxs,PinPacking{rd.availableRes.gpu,rnode.id})
    }
    taskMappings := PinPackingSlover(woods,boxs)
    ch<-taskMappings
    return taskMappings
}



func ReadFile() {
    // file, err := os.Open("./test.txt")
    file, err := os.OpenFile("./test.txt", os.O_CREATE|os.O_RDONLY, 0666)
    if err != nil {
        fmt.Println("Open file error: ", err)
        return
    }
    defer file.Close()    //关闭文件

    reader := bufio.NewReader(file)    //带缓冲区的读写
    for {
        str, err := reader.ReadString('\n')    // 以\n为分隔符来读取
        if err != nil {
            fmt.Println("read string failed, err: ", err)
            return
        }
        fmt.Println("read string is %s: ", str)
    }
}
*/