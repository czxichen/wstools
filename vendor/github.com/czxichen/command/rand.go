package command

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var rander *rand.Rand

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	rander = rand.New(src)
}

const source = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RangeString(length int) []byte {
	var str = make([]byte, 0, length+15)
	for i := 0; i < length; i++ {
		str = append(str, source[rander.Intn(36)])
	}
	return str
}

type newID struct {
	sync.Mutex
	count    int
	lastTime int64
}

var nid newID

func NewID() string {
	nid.Lock()
	newTime := time.Now().Unix()
	for {
		if newTime > nid.lastTime {
			nid.lastTime = newTime
			nid.count = 100
			break
		} else {
			nid.count += 1
			if nid.count < 1000 {
				break
			}
			newTime += 1
		}
	}
	buf := strconv.AppendInt(RangeString(5), int64(nid.count), 10)
	buf = strconv.AppendInt(buf, nid.lastTime, 10)
	nid.Unlock()
	return string(buf)
}
