package main

import (
	"fmt"
	"time"

	"github.com/aizsfgk/mdgo/net"
	"github.com/aizsfgk/mdgo/net/eventloop"
	"github.com/aizsfgk/mdgo/net/connection"
)

type echoHandler struct {
}

func (e *echoHandler) OnEventLoopInit(conn *connection.Connection) {
	fmt.Println("OnEventLoopInit")
}

func (e *echoHandler)  OnConnection(conn *connection.Connection) {
	fmt.Println("OnEventLoopInit")
}

func (e *echoHandler) OnMessage(conn *connection.Connection) {
	fmt.Println("OnEventLoopInit")
}

func (e *echoHandler) OnWriteComplete(conn *connection.Connection) {
	fmt.Println("OnEventLoopInit")
}
func (e *echoHandler) OnClose(conn *connection.Connection) {
	fmt.Println("OnEventLoopInit")
}

func main()  {



	loop, err := eventloop.New()
	if err != nil {
		fmt.Println("loop-err: ", err)
		return
	}

	var eh = &echoHandler{}

	serv, err := net.NewServer(eh, net.ReusePort(true), net.KeepAlive(10 * time.Minute))
	if err != nil{
		fmt.Println("server-err: ", err)
		return
	}

	serv.Start()
	loop.Loop()
	
}
