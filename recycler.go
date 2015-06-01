package recycler

// This package provides something similar (but not quite the same) as
// `sync.Pool` from the standard library. There are three primary
// differences:
//
// 1. `recycler` provides destructors, for clearing recycled objects and
//    breaking any cycles.
// 2. `recycler` forces the user to specify the maximum size of the free
//    list. `sync.Pool` does not give the user this level of control
//    over the free list size.
// 3. `recycler` provides an initializer (aka a constructor) which gets
//    called each time the user gets a new (or recycled) object). The
//    constructor can take user supplied parameters. Check out the tests
//    to see how this works.
//
// The motivation for this packages comes from managing complex objects
// with fixed lifecycles. The objects are complex enough that you cannot
// simply "reuse" them and expect them to work properly. They need to be
// cleared before they can be used. The number of types which need to
// have a recycler of this sort was high enough that writing a generic
// solution seems to be worth it. I hope you also find this library
// useful.

import (
	"fmt"
	"reflect"
)


type Creator func() interface{}
type Initializer func(item interface{}, params ... interface{})
type Destructor func(item interface{})

type Recycler struct {
	types map[string]chan interface{}
	creators map[string]Creator
	initializers map[string]Initializer
	destructors map[string]Destructor
}


func New() *Recycler {
	return &Recycler{
		types: make(map[string]chan interface{}),
		creators: make(map[string]Creator),
		initializers: make(map[string]Initializer),
		destructors: make(map[string]Destructor),
	}
}

func (r *Recycler) name(t reflect.Type) string {
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

func (r *Recycler) Add(t reflect.Type, c Creator, i Initializer, d Destructor, bufSize int) bool {
	name := r.name(t)
	if _, has := r.types[name]; !has {
		r.types[name] = make(chan interface{}, bufSize)
		r.creators[name] = c
		r.initializers[name] = i
		r.destructors[name] = d
	} else {
		return false
	}
	return true
}

func (r *Recycler) Get(t reflect.Type, params ...interface{}) interface{} {
	name := r.name(t)
	if _, has := r.types[name]; !has {
		panic(fmt.Errorf("Unknown type %v", t))
	}
	var item interface{}
	select {
	case item = <-r.types[name]:
	default:
		item = r.creators[name]()
	}
	r.initializers[name](item, params...)
	return item
}

func (r *Recycler) Recycle(item interface{}) {
	t := reflect.TypeOf(item)
	name := r.name(t)
	if _, has := r.types[name]; !has {
		panic(fmt.Errorf("Unknown type %v", t))
	}
	r.destructors[name](item)
	select {
	case r.types[name]<-item:
		return
	default:
		return
	}
}

