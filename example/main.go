package main

import (
	"fmt"

	"github.com/hodgesds/hydrator"
)

// A an example struct
type A struct {
	ID  int
	B   *B `hydrate:"BID"`
	BID int
	C   *C `hydrate:"GetC"`
	c   *C `hydrate:"GetC"`
}

// GetC is a method to get C
func (a *A) GetC(x interface{}) (interface{}, error) {
	println("calling GetC")
	return &C{ID: 3}, nil
}

// B is an example struct
type B struct {
	ID int
}

// C is an example struct
type C struct {
	ID int
	D  *D `hydrate:"GetD"`
}

func (c *C) GetD(x interface{}) (interface{}, error) {
	println("calling GetD")
	return &D{ID: 4}, nil
}

// D is an example struct
type D struct {
	ID int
}

func main() {
	hydrator := hydrator.NewHydrator(hydrator.Concurrency(2))

	hydrator.Finder(
		B{},
		func(obj interface{}) (interface{}, error) {
			return &B{ID: 2}, nil
		},
	)

	a := &A{
		ID:  1,
		BID: 2,
	}

	hydrator.Hydrate(a)

	fmt.Printf("a: %+v\n", a)
	fmt.Printf("a.B %+v\n", a.B)
	fmt.Printf("a.C %+v\n", a.C)
}
