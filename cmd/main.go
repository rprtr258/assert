package main

import (
	"fmt"
	"strings"

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
	return strings.Join([]string{
		fmt.Sprintf(q.Q(0, "main", "dump")),
		fmt.Sprintf(q.Q(1, "main", "dump")),
		fmt.Sprintf(q.Q(2, "main", "dump")),
		fmt.Sprintf(q.Q(3, "main", "dump")),
		fmt.Sprintf(q.Q(4, "main", "dump")),
		fmt.Sprintf(q.Q(5, "main", "dump")),
		fmt.Sprintf(q.Q(6, "main", "dump")),
		fmt.Sprintf(q.Q(7, "main", "dump")),
	}, "\n")
}

func main() {
	a := 123
	b := "hello world"
	c := 3.1415926
	d := func(n int) bool { return n > 0 }(1)
	e := []int{1, 2, 3}
	f := []byte("goodbye world")
	g := e[1:]

	fmt.Println(dump(a, b, c, d, e, f, g))
}
