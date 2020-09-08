package main

import (
	"fmt"
	"time"

	"example.com/simple-project/a"
	"example.com/simple-project/b"
)

func main() {
	fmt.Println("hello")
	a.Say()
	b.Say()
	time.Sleep(time.Second * 15)
}
