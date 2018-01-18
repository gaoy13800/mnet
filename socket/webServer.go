package socket


import (
	"log"
	"net/http"
	"fmt"
	"mnet/web"
	"mnet/conf"
)

type WebServer struct {

}


func (webs *WebServer) RunWork(){

	router :=  web.NewRouter()

	addr := fmt.Sprintf("%s:%s", "0.0.0.0", conf.Conf["web_listen_port"])

	log.Fatal(http.ListenAndServe(addr, router))
}


func NewWebServer() * WebServer{

	return &WebServer{}
}