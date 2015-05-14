package jsonlight

import (
	"io"
	"sync"
)

// it should be easy to add super functionality to existing map using simple type casting
type SynchronizedObjectWrapper struct {
	O     IObject
	Mutex sync.Mutex
}

func GetSynchronizedWrapper(o IObject) IObject {
	a := &SynchronizedObjectWrapper{
		O: o,
	}
	return a
}

//-----------------------------------------

func (this *SynchronizedObjectWrapper) ToReadonlyObject() IReadonlyObject {
	return IReadonlyObject(this)
}

func (this *SynchronizedObjectWrapper) Length() int {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Length()
}

func (this *SynchronizedObjectWrapper) Rename(oldkey, newkey string) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Rename(oldkey, newkey)
}

func (this *SynchronizedObjectWrapper) ToString(indentFactor ...int) string {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.ToString(indentFactor...)
}

func (this *SynchronizedObjectWrapper) ToByteArray(indentFactor ...int) []byte {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.ToByteArray(indentFactor...)
}

func (this *SynchronizedObjectWrapper) Write(writer *io.Writer) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.O.Write(writer)
}

func (this *SynchronizedObjectWrapper) ToMap() map[string]interface{} {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.ToMap()
}

func (this *SynchronizedObjectWrapper) ToArray(names ...string) IArray {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.ToArray(names...)
}

func (this *SynchronizedObjectWrapper) Keys() []string {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Keys()
}

func (this *SynchronizedObjectWrapper) Append(key string, value interface{}) (interface{}, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Append(key, value)
}

func (this *SynchronizedObjectWrapper) Has(key string) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Has(key)
}

func (this *SynchronizedObjectWrapper) Increment(key string) (int64, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Increment(key)
}

func (this *SynchronizedObjectWrapper) Remove(key string) interface{} {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Remove(key)
}

func (this *SynchronizedObjectWrapper) Put(key string, v interface{}) (interface{}, error) { // XJSON
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Put(key, v)
}
func (this *SynchronizedObjectWrapper) PutAll(v IObject) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.PutAll(v)
}

func (this *SynchronizedObjectWrapper) FillStruct(s interface{}) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.FillStruct(s)
}

//-------------------------------------------------------

func (this *SynchronizedObjectWrapper) Get(key string) (interface{}, bool) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Get(key)
}
func (this *SynchronizedObjectWrapper) IsNull(key string) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.IsNull(key)
}
func (this *SynchronizedObjectWrapper) GetArray(key string) (IArray, error) { // XJSON
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetArray(key)
}

func (this *SynchronizedObjectWrapper) GetBoolean(key string) (bool, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetBoolean(key)
}
func (this *SynchronizedObjectWrapper) GetString(key string) (string, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetString(key)
}
func (this *SynchronizedObjectWrapper) GetDouble(key string) (float64, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetDouble(key)
}
func (this *SynchronizedObjectWrapper) GetInt(key string) (int, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetInt(key)
}
func (this *SynchronizedObjectWrapper) GetObject(key string) (IObject, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetObject(key)
}
func (this *SynchronizedObjectWrapper) GetLong(key string) (int64, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.GetLong(key)
}

//---------------------

func (this *SynchronizedObjectWrapper) Opt(key string, defaultvalue ...interface{}) interface{} {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.Opt(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptBoolean(key string, defaultvalue ...bool) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptBoolean(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptString(key string, defaultvalue ...string) string {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptString(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptDouble(key string, defaultvalue ...float64) float64 {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptDouble(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptInt(key string, defaultvalue ...int) int {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptInt(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptArray(key string, defaultvalue ...IArray) IArray {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptArray(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptObject(key string, defaultvalue ...IObject) IObject {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptObject(key, defaultvalue...)
}
func (this *SynchronizedObjectWrapper) OptLong(key string, defaultvalue ...int64) int64 {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.O.OptLong(key, defaultvalue...)
}
