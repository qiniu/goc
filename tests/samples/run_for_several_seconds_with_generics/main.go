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
	time.Sleep(time.Second * time.Duration(mulBy(3, uint(5))))
}

type c interface {
	~int | ~uint
}

func mulBy[T c, U c](x T, y U) T {
	return x * T(y)
}
