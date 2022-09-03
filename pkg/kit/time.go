package kit

import "time"

var cstZone = time.FixedZone("CST", 8*3600)

// ParseInLocal 解析按东八区解析时间
func ParseInLocal(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, cstZone)
}
