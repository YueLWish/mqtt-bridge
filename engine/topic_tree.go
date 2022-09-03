package engine

import (
	"github.com/YueLWish/mqtt-bridge/pkg/kit"
	"github.com/pkg/errors"
	"strings"
)

type TopicFilterTree struct {
	root map[string]*Node
}

func (t *TopicFilterTree) AddFilter(topics ...string) *TopicFilterTree {
	for _, topic := range topics {
		pNode := t.root
		for _, s := range kit.SplitParticiple(topic) {
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
	tSubset := kit.SplitParticiple(topic)
	kSubset := make([]string, 0, len(tSubset))

	pNode := t.root
	for _, s := range tSubset {

		if cNode, ok := pNode["#"]; ok {
			kSubset = append(kSubset, cNode.Value)
			pNode = cNode.Child
			break
		}

		if cNode, ok := pNode["+"]; ok {
			kSubset = append(kSubset, cNode.Value)
			pNode = cNode.Child
			continue
		}

		cNode, ok := pNode[s]
		if !ok {
			break
		}

		if cNode.Value == s {
			kSubset = append(kSubset, cNode.Value)
			pNode = cNode.Child
			continue
		}

		break // 无法匹配
	}

	if len(kSubset) == 0 {
		return "", errors.Errorf("Failed to match.")
	}

	v := strings.Join(kSubset, "/")
	return strings.Replace(v, "//", "/", -1), nil
}

func NewTopicFilterTree() *TopicFilterTree {
	return &TopicFilterTree{
		root: make(map[string]*Node, 7),
	}
}
