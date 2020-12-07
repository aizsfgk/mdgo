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

func (e *echoHandler) OnMessage(conn *connection.Connection, nowUnix int64) {
	fmt.Println("OnMessage => nowUnix: ", nowUnix)

	bs := conn.InBuf.RetrieveAllAsBytes()
	fmt.Println("receive: ", string(bs))

	conn.SendByte(bs)
}

func (e *echoHandler) OnWriteComplete() {
	fmt.Println("OnWriteComplete")
}
func (e *echoHandler) OnClose() {
	fmt.Println("OnClose")
}

func main()  {


	var eh = &echoHandler{}
	serv, err := net.NewServer(eh, net.ReusePort(true), net.KeepAlive(10 * time.Minute), net.NumLoop(2))
	if err != nil{
		fmt.Println("server-err: ", err)
		return
	}

	if err := serv.Start(); err != nil {
		fmt.Println("err: ", err)
	}
	
}
