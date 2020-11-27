package main

import (
	"fmt"
	"time"

	"github.com/aizsfgk/mdgo/net"
	"github.com/aizsfgk/mdgo/net/connection"
)

type echoHandler struct {
}

func (e *echoHandler) OnEventLoopInit(conn *connection.Connection) {
	fmt.Println("OnEventLoopInit")
}

func (e *echoHandler)  OnConnection(conn *connection.Connection) {
	fmt.Println("OnConnection")
}

func (e *echoHandler) OnMessage(conn *connection.Connection) {
	fmt.Println("OnMessage")
}

func (e *echoHandler) OnWriteComplete(conn *connection.Connection) {
	fmt.Println("OnWriteComplete")
}
func (e *echoHandler) OnClose(conn *connection.Connection) {
	fmt.Println("OnClose")
}

func main()  {


	var eh = &echoHandler{}
	serv, err := net.NewServer(eh, net.ReusePort(true), net.KeepAlive(10 * time.Minute))
	if err != nil{
		fmt.Println("server-err: ", err)
		return
	}

	serv.Start()
	
}
