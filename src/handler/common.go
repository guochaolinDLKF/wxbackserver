package handler

import (
	"fmt"
	"net/http"
)

type HandlerMap map[string]func(http.ResponseWriter, *http.Request)

const lenPath = len("/")

var handlerMap HandlerMap
var handlerWCMap HandlerMap

func Init() error {

	// 各种注册
	handlerMap = make(HandlerMap)
	handlerMap["wx/wx_tokenverf"] = TokenVer
	handlerWCMap = make(HandlerMap)
	handlerWCMap[""]=CallUserMsg
	return nil
}

func ProcessPages(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.URL.Path[lenPath:]
	fmt.Println("r.Method",string(r.Method))
	if r.Method=="GET"{
		if v, ok := handlerMap[title]; ok {
			v(w, r)
		}
	}else{
		if v, ok := handlerWCMap[title]; ok {
			v(w, r)
		}
	}

}
func CallUserMsg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k,v:=range r.Form{
		fmt.Println(k+":"+v[0])
	}
}
func TokenVer(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k,v:=range r.Form{
		fmt.Println(k+":"+v[0])
	}
	signature :=r.Form["signature"]
	timestamp :=r.Form["timestamp"]
	nonce :=r.Form["nonce"]
	echostr :=r.Form["echostr"]

	fmt.Println("signature:",signature[0])
	fmt.Println("timestamp:",timestamp[0])
	fmt.Println("nonce:",nonce[0])
	fmt.Println("echostr:",echostr[0])

	fmt.Fprintf(w,string(echostr[0]))
}