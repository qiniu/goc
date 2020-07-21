package main

import (
	"example.com/multi-mains-project/foo"
	"example.com/multi-mains-project/internal"
)

func main() {
	foo.Bar1()
	foo.Bar2()
	internal.Hello()
}
