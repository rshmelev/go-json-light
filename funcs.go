package jsonlight

import (
	"bytes"
	"encoding/json"
)

func CompareObjects(oldo IObject, newo IObject) (IObject, IObject, IObject, IObject) {
	deleted, created, modified, unchanged := NewObjectOrNil(), NewObjectOrNil(), NewObjectOrNil(), NewObjectOrNil()

	left := make([]string, newo.Length())

	for k, v := range oldo.ToMap() {
		if !newo.Has(k) {
			deleted.Put(k, v)
		} else {
			left = append(left, k)
		}
	}
	for k, v := range newo.ToMap() {
		if !oldo.Has(k) {
			created.Put(k, v)
		}
	}
	for _, i := range left {
		oldval, _ := oldo.Get(i)
		newval, _ := newo.Get(i)
		joldval, _ := json.Marshal(oldval)
		jnewval, _ := json.Marshal(newval)
		if bytes.Compare(joldval, jnewval) != 0 {
			modified.Put(i, newval)
		} else {
			unchanged.Put(i, newval)
		}
	}

	return deleted, created, modified, unchanged
}
