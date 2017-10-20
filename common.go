package main
import (
	"fmt"

)
func split(sum int)(x,y int){
	x = sum*4/9
	y = sum - x
	return

}
func swap(x,y string)(string,string){
	return y,x;
}
func main(){
	var i,j int
	i = 1
	j = 2
	a,b:=swap("hello","world")
	fmt.Println(a,b)
	fmt.Println(split(9),i,j)
}