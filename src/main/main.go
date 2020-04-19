package main

import (
	"cmnfunc"
	"db"
	"fmt"
	"handler"
	"net/http"
	"util"
	"utils/log"
)


func main() {
	path2 := util.GetCurrentPath()

	//初始化配置文件
	if err := util.InitCmnConfig(path2, "cmn.json");err!=nil{
		fmt.Println(err)
		return
	}
	db.InitServerDB()
	cmnfunc.Init()
	fmt.Println("启动服务")
	handler.Init()
	http.HandleFunc("/", handler.ProcessPages)
	if err := http.ListenAndServe(":"+cmnfunc.Cfg["port"], nil); err != nil {
		log.Err(err)
	}
}
