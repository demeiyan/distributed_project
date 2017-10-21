package main
import (
	"net/rpc"
	"time"
	"net"
	"log"
	"net/http"
	"flag"
	"fmt"
)
type DispalyTime int	//定义服务标识

func (c *DispalyTime)ShowTime(arg *int, destTime *time.Time) error {
	*destTime = time.Now()
	return nil
}

func main(){
	port := flag.String("port","8080","Port")
	displayTime := new(DispalyTime)
	rpc.Register(displayTime)
	rpc.HandleHTTP()
	flag.Parse()
	listener, err := net.Listen("tcp",":"+*port)

	if err != nil{
		log.Fatal("listen err",err)
	}
	//输出本机ip
	addrs, err := net.InterfaceAddrs()
	if err !=nil{
		log.Fatal("ip err",err)
	}
	for _,address := range addrs{
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("listen ",ipnet.IP.String(),listener.Addr().String()[4:9])
			}

		}
	}
	fmt.Print("------------Server Run--------------")
	http.Serve(listener,nil)
}