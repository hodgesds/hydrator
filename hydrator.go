package hydrator

import (
	"fmt"
	"reflect"
	"sync"
)

// ErrInvalidObject is returned when a Hydrator is unable to hydrate an object.
var ErrInvalidObject = fmt.Errorf("Invalid object")

// Finder is used to find an instance.
type Finder func(interface{}) (interface{}, error)

// Opt is an option for configuring a Hydrator.
type Opt func(*Hydrator)

// Concurrency sets the concurrency of the hydrator.
func Concurrency(c int) Opt {
	return func(h *Hydrator) {
		if h.flowChan == nil {
			h.flowChan = make(chan struct{}, c)
		}
	}
}

// Tag is a Opt that is used to set the default tag that a Hydrator uses. If
// unset the default tag `hydrate` will be used.
func Tag(tag string) Opt {
	return func(h *Hydrator) {
		h.tag = tag
	}
}

// Hydrator is used to hydrate objects.
type Hydrator struct {
	sync.RWMutex
	tag      string
	finders  map[string]Finder
	flowChan chan struct{}
}

// NewHydrator returns a new Hydrator, if Concurrency is not set it defaults to 10.
func NewHydrator(opts ...Opt) *Hydrator {
	h := &Hydrator{
		tag:     "hydrate",
		finders: map[string]Finder{},
	}

	for _, opt := range opts {
		opt(h)
	}
	if h.flowChan == nil {
		h.flowChan = make(chan struct{}, 10)
	}

	return h
}

// Finder is used to set a Finder for a type.
func (h *Hydrator) Finder(obj interface{}, finder Finder) {
	objType := reflect.Indirect(reflect.ValueOf(obj)).Type()
	h.Lock()
	h.finders[objType.PkgPath()+objType.Name()] = finder
	h.Unlock()
}

// hydrationResult is the result of a hydration.
type hydrationResult struct {
	field string
	err   error
	val   interface{}
}

// Hydrate takes on object and attempts to dynamically hydrate it.
func (h *Hydrator) Hydrate(obj interface{}) error {
	var err error

	objVal := reflect.ValueOf(obj)
	objType := objVal.Type()
	indObjVal := reflect.Indirect(objVal)

	if indObjVal.Type().Kind() != reflect.Struct {
		return ErrInvalidObject
	}

	var wg sync.WaitGroup
	// make the result channel number of fields + 1 just to be safe
	resChan := make(chan hydrationResult, indObjVal.NumField()+1)

	for i := 0; i < indObjVal.NumField(); i++ {
		structField := indObjVal.Type().Field(i)
		kind := structField.Type.Kind()

		hydrateTag := structField.Tag.Get(h.tag)
		if hydrateTag == "" || hydrateTag == "-" {
			// if there is no struct tag
			continue
		}

		if structField.Anonymous || !indObjVal.CanSet() {
			err = fmt.Errorf(
				"Attempted to hydrate anonymous field %s",
				structField.Name,
			)
			break
		}

		// only hydrate pointers, slices, and arrays
		if kind != reflect.Ptr && (kind != reflect.Slice && kind != reflect.Array) {
			err = fmt.Errorf(
				"Attempted to hydrate %v field %s",
				kind,
				structField.Name,
			)
			break
		}

		// if there is a method on the struct try calling it
		_, ok := objType.MethodByName(hydrateTag)
		if ok {
			wg.Add(1)
			go func(flowChan chan struct{}) {
				defer wg.Done()
				flowChan <- struct{}{}
				vals := objVal.MethodByName(hydrateTag).Call(
					[]reflect.Value{
						objVal,
					},
				)

				var err error
				if vals[1].Interface() != nil {
					err = vals[1].Interface().(error)
				}

				resChan <- hydrationResult{
					err:   err,
					val:   vals[0].Interface(),
					field: structField.Name,
				}
				<-flowChan
			}(h.flowChan)
			continue
		}

		h.RLock()
		finder, ok := h.finders[structField.Type.Elem().PkgPath()+structField.Type.Elem().Name()]
		h.RUnlock()

		// if there is no finder then continue
		if !ok {
			continue
		}

		wg.Add(1)
		go func(flowChan chan struct{}, finder Finder) {
			defer wg.Done()
			flowChan <- struct{}{}
			val, err := finder(indObjVal.FieldByName(hydrateTag).Interface())
			resChan <- hydrationResult{
				err:   err,
				val:   val,
				field: structField.Name,
			}
			<-flowChan
		}(h.flowChan, finder)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for res := range resChan {
		if res.err != nil {
			err = res.err
			continue
		}

		field := indObjVal.FieldByName(res.field)
		if !field.CanSet() {
			err = fmt.Errorf(
				"Attempted to hydrate a private field on %T",
				obj,
			)
			continue
		}

		resVal := reflect.ValueOf(res.val)
		resType := resVal.Type()

		// recursive hydration if it is a struct
		if resType.Kind() == reflect.Ptr && resVal.Elem().Type().Kind() == reflect.Struct {
			if er := h.Hydrate(res.val); er != nil {
				err = er
				continue
			}
		}

		// recursive hydration if it is a slice or array of structs
		if (resType.Kind() == reflect.Array || resType.Kind() ==
			reflect.Slice) && ((resType.Elem().Kind() == reflect.Ptr &&
			resType.Elem().Elem().Kind() == reflect.Struct) ||
			(resType.Elem().Kind() == reflect.Struct)) {

			// hydrate slices concurrently
			var swg sync.WaitGroup
			sliceResChan := make(chan hydrationResult, resVal.Len())

			for i := 0; i < resVal.Len(); i++ {
				swg.Add(1)
				go func(flowChan chan struct{}, i int) {
					defer swg.Done()
					flowChan <- struct{}{}

					err := h.Hydrate(resVal.Index(i).Interface())

					sliceResChan <- hydrationResult{
						err: err,
					}
					<-flowChan

				}(h.flowChan, i)
			}
			go func() {
				swg.Wait()
				close(sliceResChan)
			}()
			for sliceRes := range sliceResChan {
				if sliceRes.err != nil {
					err = sliceRes.err
					continue
				}
			}

		}

		// check if the caller was lazy and messed up returning
		// the right kind
		if field.Kind() != resVal.Kind() {
			err = fmt.Errorf(
				"Attempted to hydrate %T.%s with a %T",
				obj,
				field.Type().Name(),
				res.val,
			)
			continue
		}
		field.Set(resVal)

	}

	return err
}
