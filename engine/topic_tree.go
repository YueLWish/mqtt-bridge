package engine

import (
	"github.com/pkg/errors"
	"github.com/yuelwish/mqtt-bridge/pkg/kit"
	"strings"
)

type TopicFilterTree struct {
	root map[string]*Node
}

func (t *TopicFilterTree) AddFilter(topics ...string) *TopicFilterTree {
	for _, topic := range topics {
		pNode := t.root
		for _, s := range kit.SplitTopic(topic) {
			v, ok := pNode[s]
			if !ok {
				v = &Node{Value: s, Child: make(map[string]*Node, 3)}
				pNode[s] = v
			}
			pNode = v.Child
		}
	}
	return t
}

func (t *TopicFilterTree) MathFilter(topic string) (string, error) {
	tSubset := kit.SplitTopic(topic)
	//kSubset := make([]string, 0, len(tSubset))
	var kSb strings.Builder

	pNode := t.root
	for _, s := range tSubset {

		if cNode, ok := pNode["#"]; ok {
			kSb.WriteString(cNode.Value)
			pNode = cNode.Child
			break
		}

		if cNode, ok := pNode["+"]; ok {
			kSb.WriteString(cNode.Value)
			pNode = cNode.Child
			continue
		}

		cNode, ok := pNode[s]
		if !ok {
			break
		}

		if cNode.Value == s {
			kSb.WriteString(cNode.Value)
			pNode = cNode.Child
			continue
		}

		break // 无法匹配
	}

	if kSb.Len() == 0 {
		return "", errors.Errorf("Failed to match.")
	}

	return kSb.String(), nil
}

func NewTopicFilterTree() *TopicFilterTree {
	return &TopicFilterTree{
		root: make(map[string]*Node, 7),
	}
}
