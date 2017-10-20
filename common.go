package main

import (
	"fmt"
	"runtime"
)
func sqrt(x ,z float64)float64{

	return z-(z*z-x)/(2*z)
}

func main() {
	z :=1.0
	for i:=1; i <=10;i++  {

		z=sqrt(2,z)
		if sqrt(5,z)-z < 0.1e-10 {
			break
		}
		fmt.Println(z)
	}
	fmt.Println(z)
	fmt.Print("Go runs on ")
	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("OS X.")
	case "linux":
		fmt.Println("Linux.")
	default:
		// freebsd, openbsd,
		// plan9, windows...
		fmt.Printf("%s.", os)
	}
}