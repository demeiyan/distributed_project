package main

import ("fmt"

)

func main() {
/*    fmt.Println("When`s Saturday?")
    today :=time.Now().Weekday()
    fmt.Println(time.Monday)
    switch time.Saturday{
    case today+0:
        fmt.Println("Today.")
    case today +1 :
        fmt.Println("Tomorrow.")
    case today+2:
        fmt.Println("In two days.")
    default:
        fmt.Println("Too far away.")
    }*/
    defer fmt.Println("world")
    fmt.Println("hello  ")
}