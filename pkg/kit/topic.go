package kit

import "strings"

// SplitTopic 把 topic 按 "/" 分割为数组, 数组中保留 "/"
func SplitTopic(s string) []string {

	seq := '/'
	var (
		sb strings.Builder
		a  = make([]string, 0, 7)
	)

	for _, c := range []rune(s) {
		if c == seq {
			if sb.Len() > 0 {
				a = append(a, sb.String())
				sb.Reset()
			}
			a = append(a, string(seq))
			continue
		}
		sb.WriteRune(c)
	}
	if sb.Len() > 0 {
		a = append(a, sb.String())
		sb.Reset()
	}
	return a
}
