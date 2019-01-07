package testutils

import (
	"github.com/bxcodec/faker"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func RandString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


func NewRandomMetadata() core.Metadata {
	meta := core.Metadata{}
	faker.FakeData(&meta)
	// dns label stuff
	meta.Name = "a" + RandString(6) + "a"
	meta.Namespace = "a" + RandString(6) + "a"
	return meta
}

