package kit

import "strings"

// SplitParticiple 把 topic 分词成为 数组
// topic 特殊保留 / 开头和 / 结尾
func SplitParticiple(s string) []string {

	data := make([]string, 0, 7)

	if strings.HasPrefix(s, "/") {
		data = append(data, "/")
	}

	_data := strings.FieldsFunc(s, func(r rune) bool {
		return r == '/'
	})

	data = append(data, _data...)

	if strings.HasSuffix(s, "/") {
		data = append(data, "/")
	}

	return data
}
