package main

import "fmt"
///import "flowscheduler"
import "vlog"
import "simulator"

func main(){
	fmt.Sprintf("go")
	vlog.Vlog("test.go")
	//fmt.Println( flowscheduler.Test() )
	//flowscheduler.TestGraphMain()
	//simulator.SimTestMain()
	simulator.SimJobMain()
	//flowscheduler.LabelsTest()
	fmt.Println("-----")
	simulator.FlowSimuMain()
	//fmt.Println("-----")
}