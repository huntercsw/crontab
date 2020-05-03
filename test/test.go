package main

import "fmt"

type test struct {
	m map[int]int
}

func (t *test) mSet () {
	t.m[1] = 200
}

func main() {
	t := new(test)
	t.m = make(map[int]int, 1)
	t.m[1] = 100
	t.mSet()
	fmt.Println(t)
}
