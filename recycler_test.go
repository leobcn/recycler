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

