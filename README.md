[![Build Status](https://travis-ci.org/hodgesds/hydrator.svg?branch=master)](https://travis-ci.org/hodgesds/hydrator) [![Coverage Status](https://coveralls.io/repos/github/hodgesds/hydrator/badge.svg?branch=master)](https://coveralls.io/github/hodgesds/hydrator?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/hodgesds/hydrator)](https://goreportcard.com/report/github.com/hodgesds/hydrator) [![GoDoc](https://godoc.org/github.com/hodgesds/hydrator?status.svg)](https://godoc.org/github.com/hodgesds/hydrator)

# Hydrator
Hydrator is a library that is used to hydrate go structs. It uses struct tags to control hydration.

# Example
The hydrator is invoked by using the `hydrate` tag on a struct. The tag may
either refer to another field on the struct which gets passed as an argument to
a `Finder` **or** the tag may be a method on the struct. The hydrator can work
recursively on structs as well.

```go
// A an example struct with tags
type A struct {
	ID  int
	B   *B `hydrate:"BID"`
	BID int
	C   *C `hydrate:"GetC"`
}

// GetC is a method to get C
func (a *A) GetC(x interface{}) (interface{}, error) {
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

// GetD is a method on C that gets called for hydrating D
func (c *C) GetD(x interface{}) (interface{}, error) {
	return &D{ID: 4}, nil
}

// D is an example struct
type D struct {
	ID int
}

hydrator := hydrator.NewHydrator()

hydrator.Finder(
	B{},
	func(obj interface{}) (interface{}, error) {
		return &B{ID: 2}, nil
	},
)

fmt.Printf("%+v\n", hydrator)

a := &A{
	ID:  1,
	BID: 2,
}

hydrator.Hydrate(a)

fmt.Printf("a: %+v\n", a)
fmt.Printf("a.B %+v\n", a.B)
fmt.Printf("a.C %+v\n", a.C)
```
