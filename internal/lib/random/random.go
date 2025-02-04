package random

import (
	"math/rand"
	"time"
)

func NewRandomInt() int64 {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return int64(rnd.Intn(10) + 1)
}
