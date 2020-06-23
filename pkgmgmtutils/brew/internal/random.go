package internal

import (
	"math/rand"
	"time"

	"github.com/solo-io/go-utils/pkgmgmtutils/brew/formula_updater_types"
)

func NewRandom() formula_updater_types.Random {
	return &random{}
}

type random struct{}

func (*random) Intn(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}
