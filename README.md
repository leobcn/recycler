# A Generic Object Recycler for Go

by Tim Henderson (tadh@case.edu)

## What?

This package provides something similar (but not quite the same) as
`sync.Pool` from the standard library. There are three primary
differences:

1. `recycler` provides destructors, for clearing recycled objects and
   breaking any cycles.
2. `recycler` forces the user to specify the maximum size of the free
   list. `sync.Pool` does not give the user this level of control over
   the free list size.
3. `recycler` provides an initializer (aka a constructor) which gets
   called each time the user gets a new (or recycled) object). The
   constructor can take user supplied parameters. Check out the tests to
   see how this works.

The motivation for this packages comes from managing complex objects
with fixed lifecycles. The objects are complex enough that you cannot
simply "reuse" them and expect them to work properly. They need to be
cleared before they can be used. The number of types which need to have
a recycler of this sort was high enough that writing a generic solution
seems to be worth it. I hope you also find this library useful.


## How does it work?

The basic pattern which is uses is the "buffered channel + default
select trick"

```go

var recycler chan <type>

func init() {
	recycler := make(chan <type>, 10)
}

func get<type>() {
	select {
	case item := <-recycler:
		return item
	default:
		return new<type>()
	}
}

func recycle<type>(item <type>) {
	select {
	case recycler<-item:
	default:
	}
}
```

Using that trick is fine if you have 1 or 2 types to recycle. If you
have 10 it is very frustrating to rewrite the same code over and over
again for each type. Thus, this library.

## Usage

You could install with:

    $ go get github.com/timtadh/recycler

Recycle an `*int`.

```go
package recycler

import "testing"

import (
	"reflect"
)

func TestIntPtr(t *testing.T) {
	var i *int
	R := New()
	R.Add(
		reflect.TypeOf(i),
		func() interface{} {
			t.Log("new")
			return new(int)
		},
		func(item interface{}, params ... interface{}) {
			t.Log("init")
			if len(params) > 0 {
				var i *int = item.(*int)
				*i = params[0].(int)
			}
		},
		func(item interface{}) {
			t.Log("destroy")
			var i *int = item.(*int)
			*i = 0
		},
		5,
	)
	i = R.Get(reflect.TypeOf(i), 2).(*int)
	t.Log(*i)
	if *i != 2 {
		t.Error("i != 2, i == %v", *i)
	}
	R.Recycle(i)
	i = R.Get(reflect.TypeOf(i)).(*int)
	if *i != 0 {
		t.Error("i != 0, i == %v", *i)
	}
	t.Log(*i)
}
```


