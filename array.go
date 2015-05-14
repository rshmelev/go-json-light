package jsonlight

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// re: arrays, that's not simple. because:
// 1) slices change after being appended
type JSONArray struct {
	// alternative 1
	data *[]interface{}
	// alternative 2
	m    JSONObject
	mkey string
	// alternative 3
	a    *JSONArray
	akey int

	// methods allow passing array as value
	// however that means i cannot reattach it to new parent
	// in this way the only hacky solution is to keep pointer to original struct
	originalptr *JSONArray
}

func NewArray(slicee ...*[]interface{}) *JSONArray {
	if len(slicee) > 1 {
		panic("brrr")
	}
	var a *JSONArray
	if len(slicee) == 0 {
		x := make([]interface{}, 0)
		a = &JSONArray{data: &x}
	} else {
		a = &JSONArray{data: slicee[0]}
	}
	a.originalptr = a
	return a
}

//------------------

func (this *JSONArray) Length() int {
	slice, ok := this.ToSlice()
	if !ok {
		return -1
	}
	return len(slice)
}

func (this *JSONArray) ToString(indentFactor ...int) string {
	return string(this.ToByteArray(indentFactor...))
}
func (this *JSONArray) ToByteArray(indentFactor ...int) []byte {
	a, ok := this.ToSlice()
	if !ok {
		return []byte("<EXPIREDARRAY>")
	}

	var x []byte
	if len(indentFactor) > 0 {
		x, _ = json.MarshalIndent(a, "", strings.Repeat(" ", indentFactor[0]))
	} else {
		x, _ = json.Marshal(a)
	}

	return x // return fmt.Sprintf("%+v", *(this.data))
}

func (this *JSONArray) Join(separator string) string {
	a, ok := this.ToSlice()
	if !ok {
		return "<EXPIREDARRAY>"
	}

	var buffer bytes.Buffer

	for i := 0; i < len(a); i++ {
		if i > 0 {
			buffer.WriteString(separator)
		}
		buffer.WriteString(fmt.Sprintf("%v", a[i]))
	}

	return buffer.String()
}

func (this *JSONArray) Write(writer *io.Writer) {
	enc := json.NewEncoder(*writer)
	enc.Encode(this.data)
}

func (this *JSONArray) ToSliceOrDie() []interface{} {
	x, ok := this.ToSlice()
	if !ok {
		panic("failed getting slice from Array")
	}
	return x
}

func (this *JSONArray) ToSlice() ([]interface{}, bool) {
	if this.data != nil {
		return *(this.data), true
	} else if this.a != nil {
		parentslice, ok := this.a.ToSlice()
		if !ok {
			return nil, false
		}
		elem := parentslice[this.akey] //this.a.Get(this.akey)
		//println("xxx", Dump(&elem))
		if !ok || isNil(&elem) {
			return nil, false
		}
		res, ok2 := elem.([]interface{})
		if !ok2 {
			return nil, false
		}
		return res, true
	} else {
		elem, ok := this.m[this.mkey]
		if !ok || isNil(&elem) {
			return nil, false
		}
		res, ok2 := elem.([]interface{})
		if !ok2 {
			return nil, false
		}

		return res, true
	}
}

func (this *JSONArray) Remove(index int) interface{} {
	var removed interface{}

	if this.data != nil {
		s := *(this.data)
		removed = s[index]
		s = append(s[:index], s[index+1:]...)
		*(this.data) = s
	} else if this.a != nil {
		s, ok := this.ToSlice() //this.a.Get(this.akey).([]interface{})
		if !ok {
			return nil
		}
		removed = s[index]
		s = append(s[:index], s[index+1:]...)
		this.a.Put(this.akey, s)
	} else {
		mapp := this.m.ToMap()
		s, ok := this.ToSlice() //mapp[this.mkey].([]interface{})
		if !ok {
			return nil
		}
		removed = s[index]
		s = append(s[:index], s[index+1:]...)
		mapp[this.mkey] = s
	}

	return removed
}

// TODO need to return error if wasn't able to detach well
func (this *JSONArray) DetachFromParent() {
	if this.data != nil {
		return
	} else if this.a != nil {
		slice, ok := this.ToSlice()
		this.a.Remove(this.akey)
		this.a = nil
		this.akey = -1
		if !ok {
			return
		}
		this.data = &slice
	} else {
		slice, ok := this.ToSlice()
		this.m.Remove(this.mkey)
		this.m = nil
		this.mkey = ""
		if !ok {
			return
		}
		this.data = &slice
	}
}

// can be optimized
// TODO handle slice not ok issue
func (this *JSONArray) Append(v ...interface{}) IArray {
	// println("appending ", Dump(&v[0]), " to ", fmt.Sprintf("%+v", this), this.ToString())
	a, ok := this.ToSlice()
	if !ok {
		return nil
	}

	baseindex := len(a)

	newa := a
	for i := 0; i < len(v); i++ {
		aa, isArray := v[i].(JSONArray)
		oldlen := this.Length()
		if isArray {
			(&aa).DetachFromParent()
		} else {
			paa, isArray := v[i].(*JSONArray)
			if isArray {
				(paa).DetachFromParent()
			}
		}
		newlen := this.Length()
		if oldlen != newlen {
			baseindex--
		} else {
			newa = append(newa, nil)
		}
	}

	this.updateParent(newa)

	for i := 0; i < len(v); i++ {

		this.Put(baseindex+i, v[i])
	}

	return this
}

// TODO handle expired slice
func (this *JSONArray) updateParent(v []interface{}) {
	slice := v
	if this.data != nil {
		this.data = &slice
	} else if this.a != nil {
		parentslice, ok := this.a.ToSlice()
		if !ok {
			return
		}
		parentslice[this.akey] = slice
	} else {
		this.m.Put(this.mkey, slice)
	}
}

func (this *JSONArray) Put(index int, v interface{}) (interface{}, error) { // XJSON
	a, ok := this.ToSlice()
	if !ok {
		return nil, errors.New("Expired array")
	}
	if index < 0 || index >= len(a) {
		return nil, errors.New("Array.Put: index out of range")
	}

	var prev interface{}

	switch vv := v.(type) {
	default:
		return nil, errors.New(fmt.Sprintf("Array.Put: unexpected type %T", vv))
	case JSONObject, []interface{}, bool, float32, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64, string, map[string]interface{}:
		prev = a[index]
		a[index] = v
	case *JSONArray:
		p := v.(*JSONArray)
		if p == this {
			return nil, errors.New("Array.Put: self-referencing detected")
		}
		return this.Put(index, *(p))
	case *JSONObject:
		return this.Put(index, *(v.(*JSONObject)))
	case JSONArray:
		y := v.(JSONArray).originalptr
		slice, ok := y.ToSlice()
		if !ok {
			return nil, errors.New("Array.Put: attempt to put expired array")
		}
		y.DetachFromParent()
		prev = a[index]
		a[index] = slice

		y.data = nil
		y.a = this
		y.akey = index
		y.m = nil
		y.mkey = ""
	}

	return prev, nil
}

//--------------------------------------

func (this *JSONArray) IsNull(index int) bool {
	slice, ok := this.ToSlice()
	if !ok || index < 0 || index >= len(slice) {
		return true
	}
	v := slice[index]
	return isNil(&v)
}

func (this *JSONArray) Get(index int) (interface{}, bool) {
	slice, ok := this.ToSlice()
	if !ok {
		return nil, false
	}
	if index < 0 || index >= len(slice) {
		return nil, false
	}

	return slice[index], true
}

func (this *JSONArray) GetArray(index int) (IArray, error) { // XJSON
	m, ok := this.ToSlice()
	if !ok || m == nil {
		return nil, ArrayExpiredError{}
	}
	if index < 0 || index >= len(m) {
		return nil, NotFoundError{}
	}

	v := m[index]
	_, arrok := v.([]interface{})
	if !arrok {
		return nil, TypeConvertError{}
	}
	return &JSONArray{a: this, akey: index}, nil
}

func (this *JSONArray) GetBoolean(index int) (bool, error) {
	a, ok := this.Get(index)
	if !ok {
		return false, NotFoundError{}
	}
	if isNil(&a) {
		return false, NilConvertError{}
	}
	if v, ok := a.(bool); ok {
		return v, nil
	}
	return false, TypeConvertError{}
}
func (this *JSONArray) GetString(index int) (string, error) {
	a, ok := this.Get(index)
	if !ok {
		return "", NotFoundError{}
	}
	if isNil(&a) {
		return "", NilConvertError{}
	}
	if v, ok := a.(string); ok {
		return v, nil
	}
	return "", TypeConvertError{}
}
func (this *JSONArray) GetDouble(index int) (float64, error) {
	a, ok := this.Get(index)
	if !ok {
		return 0, NotFoundError{}
	}
	if isNil(&a) {
		return 0, NilConvertError{}
	}
	if iv, ok := IntValue(a); ok {
		return float64(iv), nil
	}
	if v, ok := FloatValue(a); ok {
		return v, nil
	}
	return 0, TypeConvertError{}
}
func (this *JSONArray) GetInt(index int) (int, error) {
	long, err := this.GetLong(index)
	if err != nil {
		return 0, err
	}
	return int(long), nil
}
func (this *JSONArray) GetObject(index int) (IObject, error) {
	m, ok := this.ToSlice()
	if !ok {
		return nil, ArrayExpiredError{}
	}
	if index < 0 || index >= len(m) {
		return nil, NotFoundError{}
	}
	v := m[index]
	if !ok {
		return nil, NotFoundError{}
	}
	mm, arrok := v.(map[string]interface{})
	if !arrok {
		return nil, TypeConvertError{}
	}
	res, eee := NewObject(mm)
	return res, eee
}
func (this *JSONArray) GetLong(index int) (int64, error) {
	a, ok := this.Get(index)
	if !ok {
		return 0, NotFoundError{}
	}
	if isNil(&a) {
		return 0, NilConvertError{}
	}
	if iv, ok := IntValue(a); ok {
		return iv, nil
	}
	return 0, TypeConvertError{}
}

//-------------------------------------------------

func (this *JSONArray) Opt(index int, defaultvalue ...interface{}) interface{} {
	v, ok := this.Get(index)
	if ok {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return nil
	}
	return v
}
func (this *JSONArray) OptBoolean(index int, defaultvalue ...bool) bool {
	v, err := this.GetBoolean(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return false
	}
	return v
}
func (this *JSONArray) OptString(index int, defaultvalue ...string) string {
	v, err := this.GetString(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return ""
	}
	return v
}
func (this *JSONArray) OptDouble(index int, defaultvalue ...float64) float64 {
	v, err := this.GetDouble(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return 0
	}
	return v
}
func (this *JSONArray) OptInt(index int, defaultvalue ...int) int {
	v, err := this.GetInt(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return 0
	}
	return v
}
func (this *JSONArray) OptArray(index int, defaultvalue ...IArray) IArray {
	v, err := this.GetArray(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return NewArray()
	}
	return v
}
func (this *JSONArray) OptObject(index int, defaultvalue ...IObject) IObject {
	v, err := this.GetObject(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		o, _ := NewObject()
		return o
	}
	return v
}
func (this *JSONArray) OptLong(index int, defaultvalue ...int64) int64 {
	v, err := this.GetLong(index)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return 0
	}
	return v
}
