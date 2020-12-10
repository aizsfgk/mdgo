package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func Start(i int, msg string) {
	conn, err := net.Dial("tcp", ":19292")
	if err != nil {
		fmt.Println("Dial-err: ", err)
		return
	}
	in := msg + "; i: " + strconv.Itoa(i) + "\n"
	size := len(in)
	tmp := make([]byte, size)

	for _ = range time.NewTicker(1 * time.Second).C {
		n, err := conn.Write([]byte(in))
		if err != nil {
			fmt.Println("write-err: ", err)
		}
		if n != size {
			fmt.Println("size < : ", n)
		}

		n, err = conn.Read(tmp)
		if err != nil {
			fmt.Println("read-err: ", err)
			return
		}
		fmt.Println("read: ", string(tmp), "; n: ", n)
	}

}

func main() {
	for i := 0; i < 10024; i++ {
		go Start(i, "hello-echo")
	}

	select {}
}
