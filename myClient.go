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
	datetime :=displayTime.Format("2006-01-02 15:04:05")
	//fmt.Println(displayTime.Format("2006-01-02 15:04:05"))

	cmd := exec.Command("date","-s",fmt.Sprintf("%s",datetime))
	//cmd.Run()
	out, err := cmd.CombinedOutput()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Printf(string(out))
}