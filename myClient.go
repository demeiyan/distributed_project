package main
import(
	"net/rpc"
	"flag"
	"log"
	"time"
	"fmt"

	"os/exec"
)
func main(){
	ip := flag.String("ip","192.168.108.129","IP Adress")
	port := flag.String("port","8080","Port")
	flag.Parse()
	client, err := rpc.DialHTTP("tcp",*ip+":"+*port)
	if err !=nil{
		log.Fatal("dialing error:",err)
	}
	var arg int =0
	var displayTime time.Time
	err = client.Call("DispalyTime.ShowTime",&arg,&displayTime)
	if err != nil{
		log.Fatal("showTime error:",err)
	}
	fmt.Println(displayTime.String()[0:19])
	cmd := exec.Command("echo \"dmyan\" | sudo date ","-s","\""+displayTime.String()[0:19]+"\"")
	cmd.Run()
}