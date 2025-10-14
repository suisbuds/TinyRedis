package datastore

import (
	"github.com/suisbuds/TinyRedis/database"
	"github.com/suisbuds/TinyRedis/handler"
)

func (k *KVStore) getAsList(key string) (List, error) {
	v, ok := k.data[key]
	if !ok {
		return nil, nil
	}

	list, ok := v.(List)
	if !ok {
		return nil, handler.NewWrongTypeErrReply()
	}

	return list, nil
}

func (k *KVStore) putAsList(key string, list List) {
	k.data[key] = list
}

type List interface {
	LPush(value []byte)
	LPop(cnt int64) [][]byte
	RPush(value []byte)
	RPop(cnt int64) [][]byte
	Len() int64
	Range(start, stop int64) [][]byte
	database.CmdAdapter
}

// 双端队列
type listEntity struct {
	key    string
	buffer [][]byte
	head   int
	tail   int
	size   int
	cap    int
}

const (
	defaultCap = 16
	maxCap     = 1 << 30
)

func newListEntity(key string, elements ...[]byte) List {
	cap := defaultCap
	if len(elements) > cap {
		cap = setCap(len(elements))
	}
	l := &listEntity{
		key:    key,
		buffer: make([][]byte, cap),
		head:   0,
		tail:   len(elements),
		size:   len(elements),
		cap:    cap,
	}

	copy(l.buffer, elements)
	return l
}

// cap 是2的幂
func setCap(n int) int {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	if n > maxCap {
		return maxCap
	}
	return n
}

// 扩容
func (l *listEntity) grow() {
	newCap := l.cap * 2
	if newCap > maxCap {
		newCap = maxCap
	}
	if newCap == l.cap {
		return
	}
	newBuffer := make([][]byte, newCap)
	if l.tail > l.head {
		copy(newBuffer, l.buffer[l.head:l.tail])
	} else if l.size > 0 {
		n := copy(newBuffer, l.buffer[l.head:])
		copy(newBuffer[n:], l.buffer[:l.tail])
	}
	l.buffer = newBuffer
	l.head = 0
	l.tail = l.size
	l.cap = newCap
}

func (l *listEntity) LPush(value []byte) {
	if l.size == l.cap {
		l.grow()
	}
	l.head = (l.head - 1 + l.cap) % l.cap
	l.buffer[l.head] = value
	l.size++
}

func (l *listEntity) LPop(cnt int64) [][]byte {
	if int64(l.size) < cnt {
		return nil
	}
	res := make([][]byte, cnt)
	for i := int64(0); i < cnt; i++ {
		res[i] = l.buffer[l.head]
		l.buffer[l.head] = nil
		l.head = (l.head + 1) % l.cap
		l.size--
	}
	return res
}

func (l *listEntity) RPush(value []byte) {
	if l.size == l.cap {
		l.grow()
	}
	l.buffer[l.tail] = value
	l.tail = (l.tail + 1) % l.cap
	l.size++
}

func (l *listEntity) RPop(cnt int64) [][]byte {
	if int64(l.size) < cnt {
		return nil
	}
	res := make([][]byte, cnt)
	for i := int64(0); i < cnt; i++ {
		l.tail = (l.tail - 1 + l.cap) % l.cap
		res[cnt-1-i] = l.buffer[l.tail]
		l.buffer[l.tail] = nil
		l.size--
	}
	return res
}

func (l *listEntity) Len() int64 {
	return int64(l.size)
}

func (l *listEntity) get(idx int) []byte {
	if idx < 0 || idx >= l.size {
		return nil
	}
	i := (l.head + idx) % l.cap
	return l.buffer[i]
}

func (l *listEntity) Range(start, stop int64) [][]byte {
	if l.size == 0 {
		return nil
	}
	if stop == -1 {
		stop = int64(l.size - 1)
	}
	if start < 0 || start >= int64(l.size) {
		return nil
	}
	if stop < 0 || stop >= int64(l.size) || stop < start {
		return nil
	}
	length := stop - start + 1
	res := make([][]byte, length)
	for i := int64(0); i < length; i++ {
		res[i] = l.get(int(start + i))
	}
	return res
}

func (l *listEntity) ToCmd() [][]byte {
	args := make([][]byte, 0, 2+l.size)
	args = append(args, []byte(database.CmdTypeRPush), []byte(l.key))
	for i := 0; i < l.size; i++ {
		args = append(args, l.get(i))
	}
	return args
}
