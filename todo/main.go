package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "a"
	t := "b"
	x := strings.Join([]string{s, t}, "/")
	fmt.Println(x)
}
