package jsonlight

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// todo doc
// arrays are not detached beautifully
// no chaining
// nil cannot be type

type NotFoundError struct{}
type TypeConvertError struct{}
type NilConvertError struct{}
type ArrayExpiredError struct{}

func (a NotFoundError) Error() string     { return "Element not found" }
func (a TypeConvertError) Error() string  { return "Type convertion error" }
func (a NilConvertError) Error() string   { return "Nil convertion error" }
func (a ArrayExpiredError) Error() string { return "Array expired" }

type IBaseObject interface {
	Length() int
	ToString(indentFactor ...int) string
	ToByteArray(indentFactor ...int) []byte
	Write(writer *io.Writer)
}

type IReadonlyObject interface {
	IBaseObject

	Get(key string) (interface{}, bool)
	GetBoolean(key string) (bool, error)
	GetDouble(key string) (float64, error)
	GetInt(key string) (int, error)
	GetArray(key string) (IArray, error)
	GetObject(key string) (IObject, error)
	GetLong(key string) (int64, error)
	GetString(key string) (string, error)

	Has(key string) bool

	Opt(key string, defaultvalue ...interface{}) interface{}
	OptBoolean(key string, defaultvalue ...bool) bool
	OptDouble(key string, defaultvalue ...float64) float64
	OptInt(key string, defaultvalue ...int) int
	OptArray(key string, defaultvalue ...IArray) IArray
	OptObject(key string, defaultvalue ...IObject) IObject
	OptLong(key string, defaultvalue ...int64) int64
	OptString(key string, defaultvalue ...string) string

	ToArray(names ...string) IArray
	ToMap() map[string]interface{}
	ToReadonlyObject() IReadonlyObject

	IsNull(key string) bool

	Keys() []string
}

type IObject interface {
	IReadonlyObject

	Append(key string, value interface{}) (interface{}, error)

	Increment(key string) (int64, error)

	// returns previous value or nil if didn't exist
	Put(key string, value interface{}) (interface{}, error)
	// returns removed value
	Remove(key string) interface{}
	Rename(oldkey string, newkey string) bool

	FillStruct(s interface{}) error
}

// IObject static

type IArray interface {
	IBaseObject

	Get(index int) (interface{}, bool)
	GetBoolean(index int) (bool, error)
	GetDouble(index int) (float64, error)
	GetInt(index int) (int, error)
	GetArray(index int) (IArray, error)
	GetObject(index int) (IObject, error)
	GetLong(index int) (int64, error)
	GetString(index int) (string, error)

	Join(separator string) string
	IsNull(index int) bool

	Opt(index int, defaultvalue ...interface{}) interface{}
	OptBoolean(index int, defaultvalue ...bool) bool
	OptDouble(index int, defaultvalue ...float64) float64
	OptInt(index int, defaultvalue ...int) int
	OptArray(index int, defaultvalue ...IArray) IArray
	OptObject(index int, defaultvalue ...IObject) IObject
	OptLong(index int, defaultvalue ...int64) int64
	OptString(index int, defaultvalue ...string) string

	Put(index int, value interface{}) (interface{}, error)
	Append(values ...interface{}) IArray
	Remove(index int) interface{}

	ToSlice() ([]interface{}, bool)
}

//-------------------------------------
// helper functions

// stackoverflow guy invented this hack for some reason
// i'll better use something more lightweight until understand
// why did he add "reflect" and if it will be useful in my package
func isNil(a *interface{}) bool {
	//defer func() { recover() }()
	//return *a == nil || reflect.ValueOf(*a).IsNil()
	return *a == nil
}

func IntValue(a interface{}) (int64, bool) {
	if isNil(&a) {
		return 0, false
	}
	switch n := a.(type) {
	case int64:
		return int64(n), true
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	case float64:
		return int64(n), true
	case float32:
		return int64(n), true
	default:
		return 0, false
	}
}

func FloatValue(a interface{}) (float64, bool) {
	if isNil(&a) {
		return 0, false
	}
	switch n := a.(type) {
	case float32:
		return float64(n), true
	case float64:
		return float64(n), true
	default:
		return 0, false
	}
}

// encode to json -> decode from json.
// SLOOOOW
func StructToMap(a interface{}) (IObject, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, TypeConvertError{}
	}
	var f interface{}
	err = json.Unmarshal(b, &f)
	if res, ok := f.(map[string]interface{}); ok {
		x := NewObject(res)
		return x, nil
	}
	return nil, TypeConvertError{}
}

func Dump(a *interface{}) string {
	if isNil(a) {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", *a)
}
func Dump2(a interface{}) string {
	return fmt.Sprintf("... %-v", a)
}

//------------------------------

func GetByteContents(url string, timeout time.Duration) ([]byte, error, int) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := http.Client{
			Timeout:   timeout,
			Transport: tr,
		}
		r, err := client.Get(url)
		defer func() {
			if r != nil && r.Body != nil {
				r.Body.Close()
			}
		}()

		if err != nil {
			return nil, err, 0
		}

		x, err := ioutil.ReadAll(r.Body)

		return x, nil, r.StatusCode
	} else {
		bytes, err := ioutil.ReadFile(url)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				return nil, err, 404
			case os.IsPermission(err):
				return nil, err, 403
			default:
				return nil, err, 500
			}
		}
		return bytes, nil, 200
	}
}
