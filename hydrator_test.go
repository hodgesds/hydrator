package hydrator

import (
	"fmt"
	"testing"
)

// A an example struct
type A struct {
	ID  int
	B   *B `hydrate:"BID"`
	BID int
	C   *C `hydrate:"GetC"`
	NP  *C `hydrate:"GetNP"`
}

// GetC is a method to get C
func (a *A) GetC(x interface{}) (interface{}, error) {
	return &C{ID: 3}, nil
}

// GetNP is a non pointer receiver method to get C
func (a A) GetNP(x interface{}) (interface{}, error) {
	return &C{ID: 3}, nil
}

// B is an example struct
type B struct {
	ID int
}

// C is an example struct
type C struct {
	ID int
	D  *D   `hydrate:"GetD"`
	DD []*D `hydrate:"GetDD"`
}

func (c *C) GetD(x interface{}) (interface{}, error) {
	return &D{ID: 4}, nil
}

func (c *C) GetDD(x interface{}) (interface{}, error) {
	return []*D{&D{ID: 5}}, nil
}

// D is an example struct
type D struct {
	ID int
}

// Private is a struct to test setting private fields
type Private struct {
	ID int
	c  *C `hydrate:"GetC"`
}

// GetC is a method to get C
func (p *Private) GetC(x interface{}) (interface{}, error) {
	return &C{ID: 3}, nil
}

// NonStruct is a struct to test hydrating a non struct field
type NonStruct struct {
	ID int
	C  *int    `hydrate:"GetC"`
	D  []*int  `hydrate:"GetD"`
	E  [2]*int `hydrate:"GetE"`
}

// GetC is a method to get C
func (n *NonStruct) GetC(x interface{}) (interface{}, error) {
	res := 123
	return &res, nil
}

// GetD is a method to get D
func (n *NonStruct) GetD(x interface{}) (interface{}, error) {
	a := 1
	b := 2

	return []*int{&a, &b}, nil
}

// GetE is a method to get E
func (n *NonStruct) GetE(x interface{}) (interface{}, error) {
	a := 3
	b := 4

	return [2]*int{&a, &b}, nil
}

// NonPointer is a struct to test hydrating a non pointer field
type NonPointer struct {
	ID int
	C  int `hydrate:"GetC"`
}

// GetC is a method to get C
func (p *NonPointer) GetC(x interface{}) (interface{}, error) {
	return 3, nil
}

// HydrationError is a struct for testing hydration struct method errors
type HydrationError struct {
	ID int
	C  *C `hydrate:"GetError"`
}

// GetError is a test for when a hydration fails
func (e *HydrationError) GetError(obj interface{}) (interface{}, error) {
	return &C{ID: 1}, fmt.Errorf("Hydration error")
}

// Tagged an example struct with a custom tag
type Tagged struct {
	ID int
	C  *C `customTag:"GetC"`
}

// GetC is a method to get C
func (t *Tagged) GetC(x interface{}) (interface{}, error) {
	return &C{ID: 3}, nil
}

// RecursiveErr is a struct used to test recursive hydration errors
type RecursiveErr struct {
	ID int
	R  *RecursiveErrMethod `hydrate:"GetR"`
}

// GetR is a method to get R
func (e *RecursiveErr) GetR(x interface{}) (interface{}, error) {
	return &RecursiveErrMethod{ID: 3}, nil
}

// RecursiveErrMethod is a struct that is used to test recursive hydration
// errors
type RecursiveErrMethod struct {
	ID int
	C  *C `hydrate:"GetC"`
}

// GetC is a method to get C
func (e *RecursiveErrMethod) GetC(x interface{}) (interface{}, error) {
	return &C{ID: 3}, fmt.Errorf("Recursive error")
}

// RecursiveSliceErr is a struct used to test recursive hydration errors
type RecursiveSliceErr struct {
	ID int
	R  []*RecursiveErrMethod `hydrate:"GetR"`
}

// GetR is a method to get R
func (e *RecursiveSliceErr) GetR(x interface{}) (interface{}, error) {
	return []*RecursiveErrMethod{&RecursiveErrMethod{ID: 3}}, nil
}

func Test_Hydrator_Finder(t *testing.T) {
	h := NewHydrator(Concurrency(1))
	// add a finder for B that returns a *B with ID from BID
	h.Finder(
		B{},
		func(id interface{}) (interface{}, error) {
			idInt, ok := id.(int)
			if !ok {
				return nil, fmt.Errorf("id is non int")
			}
			return &B{ID: idInt}, nil
		},
	)

	a := &A{ID: 1, BID: 123}

	if err := h.Hydrate(a); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if a.B == nil {
		t.Errorf("Failed to hydrate B")
		t.FailNow()
	}
	if a.B.ID != a.BID {
		t.Errorf(
			"expected B to have ID %d, got %d",
			a.BID,
			a.B.ID,
		)
	}

	// check non pointer receiver hydration
	if a.NP == nil {
		t.Errorf("Failed to hydrate via non pointer receiver")
		t.FailNow()
	}

	// recursive check
	if a.C.D == nil {
		t.Errorf("Failed to recursively hydrate A.C.D")
		t.FailNow()
	}
	if a.C.D.ID != 4 {
		t.Errorf("Failed to recursively hydrate A.C.D.ID")
		t.FailNow()
	}
	if len(a.C.DD) != 1 {
		t.Errorf("Failed to recursively hydrate A.C.DD")
		t.FailNow()
	}

	// set the finder to return a B which is an invalid type
	h.Finder(
		B{},
		func(id interface{}) (interface{}, error) {
			idInt, ok := id.(int)
			if !ok {
				return nil, fmt.Errorf("id is non int")
			}
			return B{ID: idInt}, nil
		},
	)
	a = &A{ID: 1, BID: 123}

	err := h.Hydrate(a)
	if err == nil {
		t.Errorf("Hydrate set an invalid result")
		t.FailNow()
	}
}

func Test_Hydrator_no_finder(t *testing.T) {
	h := NewHydrator()
	a := &A{ID: 1, BID: 123}

	err := h.Hydrate(a)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
}

func Test_Hydrator_private(t *testing.T) {
	h := NewHydrator()

	p := Private{ID: 1}

	if err := h.Hydrate(p); err == nil {
		t.Errorf("Hydrated a private field")
		t.FailNow()
	}

	pp := &Private{ID: 1}

	if err := h.Hydrate(pp); err == nil {
		t.Errorf("Hydrated a private field")
		t.FailNow()
	}
}

func Test_Hydrator_non_struct(t *testing.T) {
	h := NewHydrator()

	n := &NonStruct{ID: 1}

	if err := h.Hydrate(n); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if n.C == nil || *n.C != 123 {
		t.Errorf("Failed to hydrate non struct, C should be 123")
		t.FailNow()
	}
	if len(n.D) != 2 || *n.D[0] != 1 || *n.D[1] != 2 {
		t.Errorf("Failed to hydrate non struct, D should be [*1,*2]")
		t.FailNow()

	}
	if len(n.E) != 2 || *n.E[0] != 3 || *n.E[1] != 4 {
		t.Errorf("Failed to hydrate non struct, E should be [*3,*4]")
		t.FailNow()

	}

	// hydrating a int should fail
	if err := h.Hydrate(1); err == nil {
		t.Errorf("Should not be able to hydrate an int")
		t.FailNow()
	}

	// hydrating a slice should fail
	if err := h.Hydrate([]*NonStruct{n}); err == nil {
		t.Errorf("Should not be able to hydrate a slice")
		t.FailNow()
	}
}

func Test_Hydrator_non_pointer(t *testing.T) {
	h := NewHydrator()

	n := NonPointer{ID: 1}

	if err := h.Hydrate(n); err == nil {
		t.Errorf("Hydrated a non pointer field")
		t.FailNow()
	}

	v := &NonPointer{ID: 1}

	if err := h.Hydrate(v); err == nil {
		t.Errorf("Hydrated a non pointer field")
		t.FailNow()
	}
}

func Test_Hydrator_error(t *testing.T) {
	h := NewHydrator()

	e := HydrationError{ID: 1}

	if err := h.Hydrate(e); err == nil {
		t.Errorf("Hydrated an error")
		t.FailNow()
	}

	v := &HydrationError{ID: 1}

	if err := h.Hydrate(v); err == nil {
		t.Errorf("Hydrated an error")
		t.FailNow()
	}
}

func Test_Hydrator_Tag(t *testing.T) {
	h := NewHydrator(Tag("customTag"))

	s := Tagged{ID: 1}

	if err := h.Hydrate(s); err != nil {
		//t.Errorf(err.Error())
		//t.FailNow()
	}

	if s.C == nil {
		//t.Errorf("Failed to hydrate custom tag")
		//t.FailNow()
	}

	p := &Tagged{ID: 1}

	if err := h.Hydrate(p); err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	if p.C == nil {
		t.Errorf("Failed to hydrate custom tag")
		t.FailNow()
	}
}

func Test_Hydrator_recursive_error(t *testing.T) {
	h := NewHydrator()

	r := RecursiveErr{ID: 1}

	if err := h.Hydrate(r); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}

	v := &RecursiveErr{ID: 1}

	if err := h.Hydrate(v); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}

	x := RecursiveSliceErr{ID: 1}

	if err := h.Hydrate(x); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}

	y := &RecursiveSliceErr{ID: 1}

	if err := h.Hydrate(y); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}
}
