package hydrator

import (
	"context"
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
func (a *A) GetC(ctx context.Context, x interface{}) (interface{}, error) {
	return &C{ID: 3}, nil
}

// GetNP is a non pointer receiver method to get C
func (a A) GetNP(ctx context.Context, x interface{}) (interface{}, error) {
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

func (c *C) GetD(ctx context.Context, x interface{}) (interface{}, error) {
	return &D{ID: 4}, nil
}

func (c *C) GetDD(ctx context.Context, x interface{}) (interface{}, error) {
	return []*D{{ID: 5}}, nil
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
func (p *Private) GetC(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
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
func (n *NonStruct) GetC(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	res := 123
	return &res, nil
}

// GetD is a method to get D
func (n *NonStruct) GetD(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	a := 1
	b := 2

	return []*int{&a, &b}, nil
}

// GetE is a method to get E
func (n *NonStruct) GetE(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
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
func (p *NonPointer) GetC(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	return 3, nil
}

// HydrationError is a struct for testing hydration struct method errors
type HydrationError struct {
	ID int
	C  *C `hydrate:"GetError"`
}

// GetError is a test for when a hydration fails
func (e *HydrationError) GetError(
	ctx context.Context,
	obj interface{},
) (interface{}, error) {
	return &C{ID: 1}, fmt.Errorf("Hydration error")
}

// Tagged an example struct with a custom tag
type Tagged struct {
	ID int
	C  *C `customTag:"GetC"`
}

// GetC is a method to get C
func (t *Tagged) GetC(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	return &C{ID: 3}, nil
}

// RecursiveErr is a struct used to test recursive hydration errors
type RecursiveErr struct {
	ID int
	R  *RecursiveErrMethod `hydrate:"GetR"`
}

// GetR is a method to get R
func (e *RecursiveErr) GetR(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	return &RecursiveErrMethod{ID: 3}, nil
}

// RecursiveErrMethod is a struct that is used to test recursive hydration
// errors
type RecursiveErrMethod struct {
	ID int
	C  *C `hydrate:"GetC"`
}

// GetC is a method to get C
func (e *RecursiveErrMethod) GetC(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	return &C{ID: 3}, fmt.Errorf("Recursive error")
}

// RecursiveSliceErr is a struct used to test recursive hydration errors
type RecursiveSliceErr struct {
	ID int
	R  []*RecursiveErrMethod `hydrate:"GetR"`
}

// GetR is a method to get R
func (e *RecursiveSliceErr) GetR(
	ctx context.Context,
	x interface{},
) (interface{}, error) {
	return []*RecursiveErrMethod{{ID: 3}}, nil
}

// NonFinderMethod is a struct used to test methods that aren't finders
type NonFinderMethod struct {
	ID int
	B  *B `hydrate:"GetB"`
}

// GetC does not implement the finder interface
func (f *NonFinderMethod) GetB(x interface{}) interface{} {
	return B{ID: 123}
}

func Test_Hydrator_Finder(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator(Concurrency(1))
	// add a finder for B that returns a *B with ID from BID
	h.Finder(
		B{},
		func(ctx context.Context, id interface{}) (interface{}, error) {
			idInt, ok := id.(int)
			if !ok {
				return nil, fmt.Errorf("id is non int")
			}
			return &B{ID: idInt}, nil
		},
	)

	a := &A{ID: 1, BID: 123}

	if err := h.Hydrate(ctx, a); err != nil {
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
		func(ctx context.Context, id interface{}) (interface{}, error) {
			idInt, ok := id.(int)
			if !ok {
				return nil, fmt.Errorf("id is non int")
			}
			return B{ID: idInt}, nil
		},
	)
	a = &A{ID: 1, BID: 123}

	err := h.Hydrate(ctx, a)
	if err == nil {
		t.Errorf("Hydrate set an invalid result")
		t.FailNow()
	}
}

func Test_Hydrator_non_finder(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()
	a := &A{ID: 1, BID: 123}

	err := h.Hydrate(ctx, a)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
}

func Test_Hydrator_private(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()

	p := Private{ID: 1}

	if err := h.Hydrate(ctx, p); err == nil {
		t.Errorf("Hydrated a private field")
		t.FailNow()
	}

	pp := &Private{ID: 1}

	if err := h.Hydrate(ctx, pp); err == nil {
		t.Errorf("Hydrated a private field")
		t.FailNow()
	}
}

func Test_Hydrator_non_struct(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()

	n := &NonStruct{ID: 1}

	if err := h.Hydrate(ctx, n); err != nil {
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
	if err := h.Hydrate(ctx, 1); err == nil {
		t.Errorf("Should not be able to hydrate an int")
		t.FailNow()
	}

	// hydrating a slice should fail
	if err := h.Hydrate(ctx, []*NonStruct{n}); err == nil {
		t.Errorf("Should not be able to hydrate a slice")
		t.FailNow()
	}
}

func Test_Hydrator_non_pointer(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()

	n := NonPointer{ID: 1}

	if err := h.Hydrate(ctx, n); err == nil {
		t.Errorf("Hydrated a non pointer field")
		t.FailNow()
	}

	v := &NonPointer{ID: 1}

	if err := h.Hydrate(ctx, v); err == nil {
		t.Errorf("Hydrated a non pointer field")
		t.FailNow()
	}
}

func Test_Hydrator_error(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()

	e := HydrationError{ID: 1}

	if err := h.Hydrate(ctx, e); err == nil {
		t.Errorf("Hydrated an error")
		t.FailNow()
	}

	v := &HydrationError{ID: 1}

	if err := h.Hydrate(ctx, v); err == nil {
		t.Errorf("Hydrated an error")
		t.FailNow()
	}
}

func Test_Hydrator_Tag(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator(Tag("customTag"))

	s := Tagged{ID: 1}

	return
	if err := h.Hydrate(ctx, s); err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	if s.C == nil {
		t.Errorf("Failed to hydrate custom tag")
		t.FailNow()
	}

	p := &Tagged{ID: 1}

	if err := h.Hydrate(ctx, p); err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	if p.C == nil {
		t.Errorf("Failed to hydrate custom tag")
		t.FailNow()
	}
}

func Test_Hydrator_recursive_error(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()

	r := RecursiveErr{ID: 1}

	if err := h.Hydrate(ctx, r); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}

	v := &RecursiveErr{ID: 1}

	if err := h.Hydrate(ctx, v); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}

	x := RecursiveSliceErr{ID: 1}

	if err := h.Hydrate(ctx, x); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}

	y := &RecursiveSliceErr{ID: 1}

	if err := h.Hydrate(ctx, y); err == nil {
		t.Errorf("Failed to error on recursive error")
		t.FailNow()
	}
}

func Test_Hydrator_non_finder_method(t *testing.T) {
	ctx := context.Background()
	h := NewHydrator()

	r := &NonFinderMethod{ID: 1}

	if err := h.Hydrate(ctx, r); err == nil {
		t.Errorf("Failed to error on non finder method")
		t.FailNow()
	}
}
