package rapidjson

import (
	"strconv"
)

type CustomJSONObject interface {
	getLength() int
	writeToBytes(b []byte) int
	Copy() CustomJSONObject
}

type JSONNull interface {
	CustomJSONObject
	isNull()
}

type JSONBool interface {
	CustomJSONObject
	Get() bool
	Set(Value bool)
}

type JSONInt interface {
	CustomJSONObject
	Get() int
	Set(Value int)
}

type JSONFloat interface {
	CustomJSONObject
	Get() float64
	Set(Value float64)
}

type JSONString interface {
	CustomJSONObject
	Get() string
	Set(Value string)
}

type JSONArray interface {
	CustomJSONObject
	Count() int
	Add(NewElement ...CustomJSONObject) int
	Insert(Index int, NewElement CustomJSONObject) int
	Remove(Index int)
	Get(Index int) CustomJSONObject
}

type JSONDictionary interface {
	CustomJSONObject
	Count() int
	Add(Key string, Value CustomJSONObject)
	Delete(Key string)
	Value(Key string) CustomJSONObject
	Keys() JSONArray
}

func JSONObjectToString(obj CustomJSONObject) string {
	if obj == nil {
		return ""
	}
	l := obj.getLength()
	m := make([]byte, l)
	obj.writeToBytes(m)
	return string(m)
}

// Null section
type jsonNull struct{}

var storageNull = jsonNull{}

func CreateNull() JSONNull {
	return &storageNull
}

func (obj *jsonNull) Copy() CustomJSONObject {
	return obj
}

func (*jsonNull) isNull() {}

func (*jsonNull) getLength() int { return 4 }

func (*jsonNull) writeToBytes(b []byte) int {
	b[0] = 'n'
	b[1] = 'u'
	b[2] = 'l'
	b[3] = 'l'
	return 4
}

// Boolean section
type jsonBool struct {
	val bool
}

func CreateBool(val bool) JSONBool {
	return &jsonBool{val: val}
}

func (obj *jsonBool) Get() bool {
	return obj.val
}

func (obj *jsonBool) Set(Value bool) {
	obj.val = Value
}

func (obj *jsonBool) getLength() int {
	if obj.val {
		return 4
	}
	return 5
}

func (obj *jsonBool) writeToBytes(b []byte) int {
	if obj.val {
		b[0] = 't'
		b[1] = 'r'
		b[2] = 'u'
		b[3] = 'e'
		return 4
	}
	b[0] = 'f'
	b[1] = 'a'
	b[2] = 'l'
	b[3] = 's'
	b[4] = 'e'
	return 5
}

func (obj *jsonBool) Copy() CustomJSONObject {
	return CreateBool(obj.val)
}

// Int section

func getIntSize(x int) int {
	p := 10
	count := 1
	for x >= p {
		count++
		p *= 10
	}
	return count
}

type jsonInt struct {
	val int
}

func CreateInt(val int) JSONInt {
	return &jsonInt{val: val}
}

func (obj *jsonInt) Get() int {
	return obj.val
}

func (obj *jsonInt) Set(Value int) {
	obj.val = Value
}

func (obj *jsonInt) getLength() int {
	if obj.val >= 0 {
		return getIntSize(obj.val)
	}
	return getIntSize(-obj.val) + 1
}

func (obj *jsonInt) writeToBytes(b []byte) int {
	var tot [20]byte
	i := 0
	var k int
	if obj.val < 0 {
		k = -obj.val
	} else {
		k = obj.val
	}
	for k >= 10 {
		q := k / 10
		tot[i] = byte(k - q*10 + '0')
		k = q
		i++
	}
	tot[i] = byte(k + '0')

	ln := i
	if obj.val < 0 {
		ln++
	}

	j := 0

	if obj.val < 0 {
		b[0] = '-'
		j++
	}
	for i >= 0 {
		b[j] = tot[i]
		j++
		i--
	}
	return ln + 1
}

func (obj *jsonInt) Copy() CustomJSONObject {
	return CreateInt(obj.val)
}

// real section

func getRealSize(x float64) int {
	return len(strconv.FormatFloat(x, 'f', -1, 64))
}

type jsonFloat struct {
	val float64
}

func CreateReal(val float64) JSONFloat {
	return &jsonFloat{val: val}
}

func (obj *jsonFloat) Get() float64 {
	return obj.val
}

func (obj *jsonFloat) Set(Value float64) {
	obj.val = Value
}

func (obj *jsonFloat) getLength() int {
	return getRealSize(obj.val)
}

func (obj *jsonFloat) writeToBytes(b []byte) int {
	str := strconv.FormatFloat(obj.val, 'f', -1, 64)
	return copy(b, []byte(str))
}

func (obj *jsonFloat) Copy() CustomJSONObject {
	return CreateReal(obj.val)
}

//  string section

const (
	dsBulk = iota
	dsSimple
	dsError
)

type jsonString struct {
	val string
	tp  int
}

func CreateString(str string) JSONString {
	return &jsonString{val: str}
}

var hex = []byte("01234567890abcdef")

func writeToBytes(str string, b []byte) int {
	src := []byte(str)
	b[0] = '"'
	offset := 1
	for _, ch := range src {
		switch ch {
		case 9:
			b[offset] = '\\'
			b[offset+1] = 't'
			offset += 2
		case 8:
			b[offset] = '\\'
			b[offset+1] = 'b'
			offset += 2
		case 10:
			b[offset] = '\\'
			b[offset+1] = 'n'
			offset += 2
		case 12:
			b[offset] = '\\'
			b[offset+1] = 'f'
			offset += 2
		case 13:
			b[offset] = '\\'
			b[offset+1] = 'r'
			offset += 2
		case '/':
			b[offset] = '\\'
			b[offset+1] = '/'
			offset += 2
		case '"':
			b[offset] = '\\'
			b[offset+1] = '"'
			offset += 2
		case '\\':
			b[offset] = '\\'
			b[offset+1] = '\\'
			offset += 2
		default:
			if ch < 0x1f {
				b[offset] = '\\'
				b[offset+1] = 'u'
				b[offset+2] = '0'
				b[offset+3] = '0'
				b[offset+4] = hex[ch>>4]
				b[offset+5] = hex[ch&0xf]
				offset += 6
			} else {
				b[offset] = ch
				offset++
			}

		}
	}
	b[offset] = '"'
	return offset + 1
}

func (obj *jsonString) writeToBytes(b []byte) int {
	return writeToBytes(obj.val, b)
}

func getLength(str string) int {
	cnt := 0
	d := []byte(str)
	for _, c := range d {
		switch c {
		case 8, 9, 10, 12, 13, '"', '\\', '/':
			cnt += 2
		default:
			if c < 0x1f {
				cnt += 6
			} else {
				cnt++
			}
		}
	}
	return cnt + 2
}

func (obj *jsonString) getLength() int {
	return getLength(obj.val)
}

func (obj *jsonString) Get() string {
	return obj.val
}
func (obj *jsonString) Set(Value string) {
	obj.val = Value
}

func (obj *jsonString) Copy() CustomJSONObject {
	return &jsonString{val: obj.val, tp: obj.tp}
}

// array section
type jsonArray struct {
	list []CustomJSONObject
	cnt  int
}

func CreateArray(initialSize int) JSONArray {
	return &jsonArray{
		list: make([]CustomJSONObject, initialSize),
		cnt:  0,
	}
}

func (obj *jsonArray) writeToBytes(b []byte) int {
	b[0] = '['
	offset := 1
	for i := 0; i < obj.cnt; i++ {
		wrt := obj.list[i].writeToBytes(b[offset:])
		offset += wrt
		if i < obj.cnt-1 {
			b[offset] = ','
			b[offset+1] = ' '
			offset += 2
		}
	}
	b[offset] = ']'
	return offset + 1

}

func (obj *jsonArray) getLength() int {
	cnt := 2
	for i := 0; i < obj.cnt; i++ {
		cnt += obj.list[i].getLength()
	}
	if obj.cnt > 1 {
		cnt += (obj.cnt - 1) * 2
	}
	return cnt
}

func (obj *jsonArray) Count() int {
	return obj.cnt
}

func (obj *jsonArray) Add(NewElements ...CustomJSONObject) int {
	i := 0
	for _, item := range NewElements {
		i = obj.Insert(obj.cnt, item)
	}
	return i
}

func (obj *jsonArray) setCapacity(newCapacity int) {
	tmp := make([]CustomJSONObject, newCapacity)
	copy(tmp, obj.list)
	obj.list = tmp
}

func (obj *jsonArray) grow() {
	var Delta int
	Cap := len(obj.list)
	if Cap > 64 {
		Delta = Cap / 4
	} else {
		if Cap > 8 {
			Delta = 16
		} else {
			Delta = 4
		}
	}
	obj.setCapacity(Cap + Delta)
}

func (obj *jsonArray) Insert(Index int, NewElement CustomJSONObject) int {
	if obj.cnt == len(obj.list) {
		obj.grow()
	}
	cnt := obj.cnt
	if Index < 0 || Index >= cnt {
		obj.list[cnt] = NewElement
		obj.cnt++
		return cnt
	}
	copy(obj.list[Index+1:], obj.list[Index:])
	obj.list[Index] = NewElement
	obj.cnt++
	return Index
}

func (obj *jsonArray) Remove(Index int) {
	if Index < 0 || Index >= obj.cnt {
		return
	}
	copy(obj.list[Index:], obj.list[Index+1:])
	obj.cnt--
}

func (obj *jsonArray) Get(Index int) CustomJSONObject {
	if Index < 0 || Index >= obj.cnt {
		return CreateNull()
	}
	return obj.list[Index]
}

func (obj *jsonArray) Copy() CustomJSONObject {
	tmp := CreateArray(obj.cnt)
	for i := 0; i < obj.cnt; i++ {
		tmp.Add(obj.list[i].Copy())
	}
	return tmp
}

// dictionary section

type jsonDictionary struct {
	dict map[string]CustomJSONObject
}

func CreateDictionary(initialSize int) JSONDictionary {
	return &jsonDictionary{
		dict: make(map[string]CustomJSONObject, initialSize),
	}
}

func (obj *jsonDictionary) writeToBytes(b []byte) int {
	b[0] = '{'
	offset := 1
	i := 0
	maplen := len(obj.dict)
	for k, v := range obj.dict {
		off := writeToBytes(k, b[offset:])
		offset += off
		b[offset] = ':'
		offset++
		off = v.writeToBytes(b[offset:])

		offset += off
		if i < maplen-1 {
			b[offset] = ','
			b[offset+1] = ' '
			offset += 2
		}
		i++
	}
	b[offset] = '}'
	return offset + 1

}

func (obj *jsonDictionary) getLength() int {
	cnt := 2
	for k, v := range obj.dict {
		cnt += getLength(k) + 1 + v.getLength()
	}
	if len(obj.dict) > 1 {
		cnt += (len(obj.dict) - 1) * 2
	}
	return cnt
}

func (obj *jsonDictionary) Count() int {
	return len(obj.dict)
}
func (obj *jsonDictionary) Add(Key string, Value CustomJSONObject) {
	_, isNull := Value.(JSONNull)
	if isNull {
		obj.Delete(Key)
		return
	}
	obj.dict[Key] = Value
}

func (obj *jsonDictionary) Delete(Key string) {
	delete(obj.dict, Key)
}

func (obj *jsonDictionary) Value(Key string) CustomJSONObject {
	Value, ok := obj.dict[Key]
	if !ok {
		return CreateNull()
	}
	return Value
}

func (obj *jsonDictionary) Keys() JSONArray {
	arr := CreateArray(len(obj.dict))
	for key := range obj.dict {
		arr.Add(CreateString(key))
	}
	return arr
}

func (obj *jsonDictionary) Copy() CustomJSONObject {
	l := len(obj.dict)
	tmp := CreateDictionary(l)
	for k, v := range obj.dict {
		tmp.Add(k, v.Copy())
	}
	return tmp
}
