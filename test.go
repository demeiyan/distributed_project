package main

import (
	"time"
	"fmt"
	"os/exec"
)
type point struct {
	x, y int
}

func main() {

	datetime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(fmt.Sprintf("\"%s\"",datetime))
	cmd :=exec.Command("date","-s",fmt.Sprintf("%s",datetime))
	out, err :=cmd.CombinedOutput()
	if err !=nil{
		fmt.Println(err)
	}
	fmt.Println(string(out))
	p := point{1, 2}
	fmt.Printf("%v\n", p)
}