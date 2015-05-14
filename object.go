package jsonlight

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// it should be easy to add super functionality to existing map using simple type casting
type JSONObject map[string]interface{}

func NewEmptyObject() IObject {
	x := JSONObject(make(map[string]interface{}))
	return IObject(&x)
}

func NewObjectOrNil(string_or_bytes_or_map_or_struct ...interface{}) IObject {
	o, _ := NewObject(string_or_bytes_or_map_or_struct...)
	return o
}

func NewObjectOrDie(string_or_bytes_or_map_or_struct ...interface{}) IObject {
	o, e := NewObject(string_or_bytes_or_map_or_struct...)
	if e != nil {
		panic(e)
	}
	if o == nil {
		panic("nil object created")
	}
	return o
}

func NewObject(string_or_bytes_or_map_or_struct ...interface{}) (IObject, error) {
	paramsLen := len(string_or_bytes_or_map_or_struct)
	if paramsLen > 0 {
		p := string_or_bytes_or_map_or_struct[0]
		v := reflect.ValueOf(p)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return NewObject()
			}
			p = v.Elem().Interface()
			v = reflect.ValueOf(p)
		}
		switch v.Kind() {
		case reflect.Map:
			m, ok := p.(map[string]interface{})
			if !ok {
				return nil, errors.New("map should have type: map[string]interface{}")
			}
			return NewObjectFromMap(m), nil
		case reflect.String:
			return NewObjectFromString(v.String())
		case reflect.Struct:
			return NewObjectFromStruct(v.Interface())
		case reflect.Slice:
			m, ok := p.([]byte)
			if !ok {
				fmt.Println(v.Type(), v.Kind(), v)
				return nil, errors.New("slice should have type: []byte")
			}
			return NewObjectFromBytes(m)
		}
		return nil, errors.New("unsupported input type")
	} else {
		return NewEmptyObject(), nil
	}
}

func NewObjectFromMap(mapp ...map[string]interface{}) IObject {
	if len(mapp) > 1 {
		panic("brrr")
	}

	if len(mapp) == 1 {
		if mapp[0] == nil {
			return nil
		}
		x := JSONObject(mapp[0]) //mapp.(JSONObject)
		return IObject(&x)
	} else {
		return NewEmptyObject()
	}
}

func NewObjectFromFile(url string, timeout time.Duration) (IObject, error, int) {

	content, err, code := GetByteContents(url, timeout)
	if err != nil {
		return nil, err, code
	}

	o, err := NewObjectFromBytes(content)
	return o, err, code
}

func NewObjectFromString(str string) (IObject, error) {
	return NewObjectFromBytes([]byte(str))
}
func NewObjectFromStruct(a interface{}) (IObject, error) {
	b, e := json.Marshal(a)
	if e != nil {
		return nil, e
	}
	return NewObjectFromBytes(b)
}
func NewObjectFromStructOrDie(a interface{}) IObject {
	b, e := json.Marshal(a)
	if e != nil {
		panic("cannot make json from object")
	}
	r, e2 := NewObjectFromBytes(b)
	if e2 != nil {
		panic("cannot create object from bytes: " + string(b) + " / " + e2.Error())
	}
	return r
}
func NewObjectFromBytes(bytes []byte) (IObject, error) {
	var res interface{}
	if err := json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}
	res2, ok := res.(map[string]interface{})
	if !ok {
		return nil, TypeConvertError{}
	}

	return NewObject(res2)
}

//-----------------------------------------

func (this *JSONObject) Length() int {
	return len(*this)
}

func (this *JSONObject) ToReadonlyObject() IReadonlyObject {
	return IReadonlyObject(this)
}

func (this *JSONObject) Rename(oldkey, newkey string) bool {
	x, ok := this.Get(oldkey)
	if !ok {
		return false
	}
	this.Put(newkey, x)
	this.Remove(oldkey)
	return true
}

func (this *JSONObject) ToString(indentFactor ...int) string {
	return string(this.ToByteArray(indentFactor...))
}
func (this *JSONObject) ToByteArray(indentFactor ...int) []byte {
	//return fmt.Sprintf("%+v", *this)

	thismap := this.ToMap()
	var x []byte
	if len(indentFactor) > 0 {
		x, _ = json.MarshalIndent(thismap, "", strings.Repeat(" ", indentFactor[0]))
	} else {
		x, _ = json.Marshal(thismap)
	}

	return x
}

func (this *JSONObject) Write(writer *io.Writer) {
	enc := json.NewEncoder(*writer)
	enc.Encode(*this)
}

func (this *JSONObject) ToMap() map[string]interface{} {
	return map[string]interface{}(*this)
}

func (this *JSONObject) ToArray(names ...string) IArray {
	m := this.ToMap()
	res := make([]interface{}, 0)
	for _, name := range names {
		res = append(res, m[name])
	}

	return NewArray(&res)
}

func (this *JSONObject) Keys() []string {
	m := this.ToMap()
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

func (this *JSONObject) Append(key string, value interface{}) (interface{}, error) {
	arr, err := this.GetArray(key)
	if err != nil {
		return nil, err
	}
	if arr == nil {
		this.ToMap()[key] = []interface{}{}
		arr, _ = this.GetArray(key)
	}
	return arr.Append(value), nil
}

func (this *JSONObject) Has(key string) bool {
	_, ok := this.ToMap()[key]
	return ok
}

func (this *JSONObject) Increment(key string) (int64, error) {
	m := this.ToMap()
	v, ok := m[key]
	if !ok {
		base := 1
		m[key] = base
		return int64(base), nil
	}
	intval, ok := IntValue(v)
	if !ok {
		return 0, errors.New("Object.Increment: non-int value already exists under key")
	}
	intval++
	m[key] = intval
	return intval, nil
}

func (this *JSONObject) Remove(key string) interface{} {
	thismap := this.ToMap()
	removed, exists := thismap[key]
	delete(thismap, key)
	if exists {
		return removed
	}
	return nil
}

func (this *JSONObject) putArray(key string, a *JSONArray) (interface{}, error) { // XJSON
	thismap := this.ToMap()
	var prev interface{}
	prevexists := false

	a.DetachFromParent()
	slice, ok := a.ToSlice()
	if !ok {
		return nil, errors.New("Array.Put: attempt to put expired array")
	}
	prev, prevexists = thismap[key]
	thismap[key] = slice

	a.data = nil
	a.a = nil
	a.akey = -1
	a.m = *this
	a.mkey = key
	// ensure that original array is updated
	// this will not help to avoid some sad expiration issues
	// but at least will make this module more secure...
	if a.originalptr != a {
		b := a.originalptr
		b.data = nil
		b.a = nil
		b.akey = -1
		b.m = *this
		b.mkey = key
	}

	if prevexists {
		return prev, nil
	}
	return nil, nil
}

func (this *JSONObject) PutAll(o IObject) error {
	if o == nil {
		return errors.New("PutAll called with nil param")
	}
	m := o.ToMap()
	for k, v := range m {
		if _, e := this.Put(k, v); e != nil {
			return e
		}
	}
	return nil
}

func (this *JSONObject) Put(key string, v interface{}) (interface{}, error) { // XJSON
	thismap := this.ToMap()
	var prev interface{}
	prevexists := false

	switch vv := v.(type) {
	default:
		return nil, errors.New(fmt.Sprintf("Object.Put: unexpected type %T", vv))
	case JSONObject, []interface{}, int, bool, float32, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64, string, map[string]interface{}:
		prev, prevexists = thismap[key]
		thismap[key] = v
	case *JSONArray:
		return this.putArray(key, v.(*JSONArray)) //this.Put(key, *(v.(*JSONArray)))
	case *JSONObject:
		return this.Put(key, *(v.(*JSONObject)))
	case JSONArray:
		return this.putArray(key, v.(JSONArray).originalptr)
	}

	if prevexists {
		return prev, nil
	}
	return nil, nil
}

func (this *JSONObject) FillStruct(s interface{}) error {
	ba := this.ToByteArray()
	if ba == nil {
		return TypeConvertError{}
	}
	return json.Unmarshal(ba, s)
}

//-------------------------------------------------------

func (this *JSONObject) Get(key string) (interface{}, bool) {
	v, ok := this.ToMap()[key]
	return v, ok
}
func (this *JSONObject) IsNull(key string) bool {
	v, ok := this.ToMap()[key]
	return !ok || isNil(&v)
}
func (this *JSONObject) GetArray(key string) (IArray, error) { // XJSON
	m := this.ToMap()
	v, ok := m[key]
	if !ok {
		return nil, NotFoundError{}
	}
	_, arrok := v.([]interface{})
	if !arrok {
		return nil, TypeConvertError{}
	}
	return &JSONArray{m: m, mkey: key}, nil
}

func (this *JSONObject) GetBoolean(key string) (bool, error) {
	a, ok := this.Get(key)
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
func (this *JSONObject) GetString(key string) (string, error) {
	a, ok := this.Get(key)
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
func (this *JSONObject) GetDouble(key string) (float64, error) {
	a, ok := this.Get(key)
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
func (this *JSONObject) GetInt(key string) (int, error) {
	long, err := this.GetLong(key)
	if err != nil {
		return 0, err
	}
	return int(long), nil
}
func (this *JSONObject) GetObject(key string) (IObject, error) {
	m := this.ToMap()
	v, ok := m[key]
	if !ok {
		return nil, NotFoundError{}
	}
	mm, arrok := v.(map[string]interface{})
	if !arrok {
		return nil, TypeConvertError{}
	}

	return NewObject(mm)
}
func (this *JSONObject) GetLong(key string) (int64, error) {
	a, ok := this.Get(key)
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

//---------------------

func (this *JSONObject) Opt(key string, defaultvalue ...interface{}) interface{} {
	v, ok := this.Get(key)
	if ok {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return nil
	}
	return v
}
func (this *JSONObject) OptBoolean(key string, defaultvalue ...bool) bool {
	v, err := this.GetBoolean(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return false
	}
	return v
}
func (this *JSONObject) OptString(key string, defaultvalue ...string) string {
	v, err := this.GetString(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return ""
	}
	return v
}
func (this *JSONObject) OptDouble(key string, defaultvalue ...float64) float64 {
	v, err := this.GetDouble(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return 0
	}
	return v
}
func (this *JSONObject) OptInt(key string, defaultvalue ...int) int {
	v, err := this.GetInt(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return 0
	}
	return v
}
func (this *JSONObject) OptArray(key string, defaultvalue ...IArray) IArray {
	v, err := this.GetArray(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return nil
	}
	return v
}
func (this *JSONObject) OptObject(key string, defaultvalue ...IObject) IObject {
	v, err := this.GetObject(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return nil
	}
	return v
}
func (this *JSONObject) OptLong(key string, defaultvalue ...int64) int64 {
	v, err := this.GetLong(key)
	if err != nil {
		if len(defaultvalue) > 0 {
			return defaultvalue[0]
		}
		return 0
	}
	return v
}
