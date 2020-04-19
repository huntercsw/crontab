package main

import "fmt"

type st struct {
	stSub
	isSub
}

type stSub struct {

}

type isSub interface {
	Get()
}

func (s *stSub) f(args ...interface{}) {
	fmt.Println(args)
	fmt.Println(args...)
}

func main() {
	s := new(st)
	s.f(1, 2, 3)
}
