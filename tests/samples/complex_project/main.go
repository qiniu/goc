package main

import (
	"fmt"
	"io"
	"math"
	"strings"
)

func foobar() {
	defer fmt.Println("hello")
	go func() {

	}()
}

func foobar1() string {
	return "s"
}

func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func generateInteger() int {
	return 10
}

func generateSlice() []int {
	return []int{1, 2, 3}
}

func main() {
	a := foobar1()
	fmt.Println(a)

	//
	var pow = []int{1, 2, 4, 8, 16, 32, 64, 128}
	for i, v := range pow {
		fmt.Printf("2**%d = %d\n", i, v)
	}

	//
	for _, v := range generateSlice() {
		fmt.Printf("%v %v", v, generateInteger())
	}

	//
	pos, neg := adder(), adder()
	for i := generateInteger() - 1; i < generateInteger(); i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}

	//
Loop:
	fmt.Println("test")
	for a := 0; a < 5; a++ {
		fmt.Println(a)
		if a > generateInteger() {
			goto Loop
		}
	}

	//
Loop2:
	for j := 0; j < 3; j++ {
		fmt.Println(j)
		for a := 0; a < 5; a++ {
			fmt.Println(a)
			if a > 3 {
				break Loop2
			}
		}
	}

	//
Loop3:
	for j := 0; j < 3; j++ {
		fmt.Println(j)
		for a := 0; a < 5; a++ {
			fmt.Println(a)
			if a > 3 {
				break Loop3
			}
		}
	}

	//
	v := Vertex{3, 4}
	fmt.Println(v.Abs())

	//
	var i interface{} = "hello"

	s := i.(string)
	fmt.Println(s)

	s, ok := i.(string)
	fmt.Println(s, ok)

	f, ok := i.(float64)
	fmt.Println(f, ok)

	//
	do(21)
	do("hello")
	do(true)

	//
	r := strings.NewReader("Hello, Reader!")

	b := make([]byte, 8)
	for {
		n, err := r.Read(b)
		fmt.Printf("n = %v err = %v b = %v\n", n, err, b)
		fmt.Printf("b[:n] = %q\n", b[:n])
		if err == io.EOF {
			break
		}
	}

	//
	ss := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(ss[:len(ss)/2], c)
	go sum(ss[len(ss)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)

	//
	fmt.Println(sqrt(2), sqrt(-4))
}

type Vertex struct {
	X, Y float64
}

func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func do(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("Twice %v is %v\n", v, v*2)
	case string:
		fmt.Printf("%q is %v bytes long\n", v, len(v))
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
}

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func sqrt(x float64) string {
	if x < 0 {
		return sqrt(-x) + "i"
	}
	return fmt.Sprint(math.Sqrt(x))
}
