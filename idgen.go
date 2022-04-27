package crex

import (
	"strconv"

	"github.com/WymA/crex/utils"
)

var idGen *utils.IdGenerate

func SetIdGenerate(g *utils.IdGenerate) {
	idGen = g
}

func GenOrderId() string {
	id := idGen.Next()
	return strconv.Itoa(int(id))
}
