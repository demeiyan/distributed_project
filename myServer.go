package main

import (
	"net/rpc"
	"log"
	"net"
	"runtime"
	"time"
	"net/http"
)

type DispalyTime int	//定义服务标识
var password string = "root"
func (c *DispalyTime)ShowTime(key *string, destTime *time.Time) error {
	if password == *key{
		*destTime = time.Now()
	}
	return nil
}
func main()  {
	runtime.GOMAXPROCS(4)
	arith := new(DispalyTime)
	server := rpc.NewServer()
	log.Printf("Register service:%v\n", arith)
	server.Register(arith)
	log.Printf("Listen tcp on port %d\n", 8080)
	l, e := net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	log.Println("Ready to accept connection...")
	conCount := 0
	 func() {
		for {
			//conn, err := l.Accept()
/*			if err != nil {
				log.Fatal("Accept Error:,", err)
				continue
			}*/
			conCount++
			log.Printf("Receive Client Connection %d\n", conCount)
			http.Serve(l,nil)
			//go server.ServeConn(conn)
		}
	}()
}