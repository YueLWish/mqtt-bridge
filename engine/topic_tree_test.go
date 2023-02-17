package engine

import (
	"testing"
)

func TestTopicFilterTreeV1(t *testing.T) {
	tfTree := NewTopicFilterTree()
	tfTree.AddFilter("/abc/#")

	topicMap := map[string]bool{
		"/abc/123":  true,
		"//abc/123": false,
		"abc/#":     false,
	}

	for topic, want := range topicMap {
		got, err := tfTree.MathFilter(topic)
		if !(err == nil && want) {
			t.Errorf("MathFilter() = %v, want %v", got, want)
		}
	}

}

func TestTopicFilterTree_MathFilter(t *testing.T) {
	tfTree := NewTopicFilterTree()
	tfTree.AddFilter("/abc/#")

	type fields struct {
		root map[string]*Node
	}
	type args struct {
		topic string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "",
			fields:  fields{},
			args:    args{},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			t := &TopicFilterTree{
				root: tt.fields.root,
			}
			got, err := t.MathFilter(tt.args.topic)
			if (err != nil) != tt.wantErr {
				t1.Errorf("MathFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("MathFilter() got = %v, want %v", got, tt.want)
			}
		})
	}
}
