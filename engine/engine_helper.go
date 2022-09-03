package engine

import (
	"github.com/pkg/errors"
)

func NewEngineHelper() *EngineHelper {
	e := new(EngineHelper)
	e.filterMap = make(map[string]*TopicFilter, 10)
	e.cliAddrMap = make(map[string]*MqttAddress, 7)
	e.routers = make([]*Router, 0, 7)
	return e
}

type EngineHelper struct {
	cliAddrMap map[string]*MqttAddress // 客户端 key: tag Value: 客户端连接信息
	filterMap  map[string]*TopicFilter // topic filter key: tag Value: 要操作的 topic
	routers    []*Router
}

func (e *EngineHelper) AddTopicFilter(tag string, qos byte, filter ...string) error {
	v, ok := e.filterMap[tag]
	if ok {
		return errors.Errorf("already exists tag: %v", tag)
	} else {

		e.filterMap[tag] = v
	}

	e.filterMap[tag] = &TopicFilter{
		Tag:    tag,
		Qos:    qos,
		Filter: filter,
	}
	return nil
}

type MqttOption func(*MqttAddress)

func (m *EngineHelper) AddClient(tag string, address string, opts ...MqttOption) error {
	if _, ok := m.cliAddrMap[tag]; ok {
		return errors.Errorf("already exists tag: %v", tag)
	}
	var it = &MqttAddress{
		Address: address,
	}
	for _, opt := range opts {
		opt(it)
	}
	m.cliAddrMap[tag] = it
	return nil
}

func (e *EngineHelper) AddRouting(fromTags, toTags, topicTags []string) {
	e.routers = append(e.routers, &Router{
		FromTags:  fromTags,
		ToTags:    toTags,
		TopicTags: topicTags,
	})
}

func (e *EngineHelper) BuildEngine() (*Engine, error) {
	var (
		cliSubMap  = make(map[string][]*SubTopic, 7) // 客户端tag : 订阅的topicFilter
		filterTree = NewTopicFilterTree()            // filterTree 用于匹配接收的topic
		toTopicMap = make(map[string][]string, 7)    // 来源tag和topicFilter 对应接收的 tags
	)

	// 构建 客户端tag : 订阅的topicFilter
	for _, router := range e.routers {
		var (
			fromTags  = router.FromTags
			topicTags = router.TopicTags
		)

		for _, cTag := range fromTags {
			_, ok := cliSubMap[cTag]
			if !ok {
				cliSubMap[cTag] = make([]*SubTopic, 0, 7)
			}

			for _, tTag := range topicTags {
				topicFilter, ok := e.filterMap[tTag]
				if !ok {
					return nil, errors.Errorf("")
				}

				filters := make([]*SubTopic, 0, len(topicFilter.Filter))
				for _, topic := range topicFilter.Filter {
					filters = append(filters, &SubTopic{
						Topic: topic,
						Qos:   topicFilter.Qos,
					})
				}

				cliSubMap[cTag] = append(cliSubMap[cTag], filters...)
			}
		}
	}

	// 构建 filterTree
	// 按 / 作为分割, 将得到的数组组装成树形结构
	for _, router := range e.routers { // 1. 循环 处理路由
		for _, topicTag := range router.TopicTags { // 2. 循环路由 中 topicTag
			tFilters, ok := e.filterMap[topicTag]
			if !ok {
				return nil, errors.Errorf("routing 未知的 topicTag %s", topicTag)
			}
			filterTree.AddFilter(tFilters.Filter...)
		}
	}

	// 构建 确定来源和filter 对应的 目标 tags
	for _, router := range e.routers { // 1. 循环路由
		var (
			fromTags  = router.FromTags
			toTags    = router.ToTags
			topicTags = router.TopicTags
		)
		for _, topicTag := range topicTags { // 2. 循环路由中的 topicTag
			topicFilter, ok := e.filterMap[topicTag]
			if !ok {
				return nil, errors.Errorf("routing 未知的 topicTag %s", topicTag)
			}

			for _, topic := range topicFilter.Filter { // 3. 循环 topicTag 中的 filter
				for _, fTag := range fromTags { // 4. 为每一个fromTag 创建 key, 并把路由所有的 toTag 作为 Value
					key := fTag + "-" + topic
					v, ok := toTopicMap[key]
					if !ok {
						v = make([]string, 0, 7)
						toTopicMap[key] = v
					}
					toTopicMap[key] = append(toTopicMap[key], toTags...)
				}
			}
		}
	}

	return &Engine{
		cliAddrMap:  e.cliAddrMap,
		cliSubMap:   cliSubMap,
		filterTree:  filterTree,
		toTopicMap:  toTopicMap,
		MessageChan: make(chan *Message, 1024),
	}, nil
}
