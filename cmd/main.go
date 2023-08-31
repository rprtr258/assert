package main

import (
	"fmt"

	"github.com/rprtr258/assert/q"
)

func dump(
	a int,
	b string,
	c float64,
	d bool,
	e []int,
	f []byte,
	g []int,
) string {
	return fmt.Sprint(q.Q("main", "dump"))
}

func main() {
	e := []int{1, 2, 3}

	fmt.Println(dump(
		123,
		"hello world",
		3.1415926,
		func(n int) bool { return n > 0 }(1),
		[]int{1, 2, 3},
		[]byte("goodbye world"),
		e[1:],
	))
}
