package kit

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

func NewTraceId() string {
	strId := strconv.FormatInt(rand.Int64(), 36)
	if len(strId) == 12 {
		return strId
	}
	if len(strId) < 12 {
		return fmt.Sprintf("%012s", strId)
	}
	return strId[:12]
}
