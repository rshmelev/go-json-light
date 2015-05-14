package jsonlight

import (
	"fmt"
	"testing"
)

type TestStruct struct {
	A int `json:"umm"`
}

func TestSomething(t *testing.T) {
	o, _ := NewObjectFromString(`{"umm":33}`)
	fmt.Printf("%+v\n", o.ToMap())
	x := &TestStruct{}
	fmt.Printf("%+v\n", x)
	o.Put("A", 10)
	o.FillStruct(x)
	fmt.Printf("%+v\n", x)
	m := StructToMapOrDie(x).ToMap()
	m["umm"] = 10
	fmt.Printf("%+v\n", m)
	NewObject(m).FillStruct(x)
	fmt.Printf("%+v\n", x)
}
