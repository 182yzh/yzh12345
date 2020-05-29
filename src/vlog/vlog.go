package  vlog

import (
	"fmt"
	"log"
	"os"
)

var vlogger *log.Logger
var dlogger *log.Logger
var temlogger  *log.Logger

func Init(filename string)*log.Logger{
	logfile,err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
    if err!= nil {
        fmt.Println(err)
    }
    lg := log.New(logfile, "Log: ", log.Ldate | log.Ltime)
    return lg
}

func Dlog(str string) {
	if dlogger == nil{	
		dlogger = Init("debug_log.log")
	}
	dlogger.Println(str)
}
func Vlog(str string){
	if vlogger == nil{
		vlogger = Init("logfile.log")
	}
	//vlogger.Println(str)
}

func Temlog(str string) {
	if temlogger == nil{	
		temlogger = Init("temlog.log")
	}
	temlogger.Printf(str)
}
