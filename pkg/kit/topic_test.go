package kit

import (
	"reflect"
	"testing"
)

func TestSplitTopic(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []string
	}{
		{name: "1", args: "/a/b/c", want: []string{"/", "a", "/", "b", "/", "c"}},
		{name: "2", args: "/ab/b/", want: []string{"/", "ab", "/", "b", "/"}},
		{name: "3", args: "/a//c123", want: []string{"/", "a", "/", "/", "c123"}},
		{name: "4", args: "//abc//cde", want: []string{"/", "/", "abc", "/", "/", "cde"}},
		{name: "5", args: "/ab/b//", want: []string{"/", "ab", "/", "b", "/", "/"}},
		{name: "6", args: "ab/b", want: []string{"ab", "/", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitTopic(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitTopic() = %v, want %v", got, tt.want)
			}
		})
	}
}
