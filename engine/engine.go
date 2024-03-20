package engine

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/yuelwish/mqtt-bridge/pkg/xmqtt"
	"log"
	"runtime"
)

const (
	clientId = "mqtt-bridge"
)

type Engine struct {
	cliAddrMap map[string]*MqttAddress // clientTag : client address
	cliSubMap  map[string][]*SubTopic  // clientTag : 对应订阅的 topicFilter
	filterTree *TopicFilterTree        // filter 按 / 分割生成树形结构
	toTopicMap map[string][]string     // fromTag+topicFilter : toTags

	MessageChan chan *Message          // 处理MQTT消息的队列
	cliConnMap  map[string]mqtt.Client // clientTag : mqtt.client
}

func (e *Engine) Dial() error {
	// 初始化 tag : 连接
	e.cliConnMap = make(map[string]mqtt.Client, len(e.cliAddrMap))

	for tag, addr := range e.cliAddrMap {
		client, err := xmqtt.Init(clientId, addr.Address, func(opt *mqtt.ClientOptions) {
			if addr.UserName != "" {
				opt.SetUsername(addr.UserName)
			}
			if addr.Password != "" {
				opt.SetPassword(addr.Password)
			}
		})
		if err != nil {
			return errors.WithMessagef(err, "mqtt init failed: %#v", addr)
		}
		e.cliConnMap[tag] = client
	}
	return nil
}

func (e *Engine) handlerMessage(ctx context.Context) {
	gPool, _ := ants.NewPool(runtime.NumCPU() * 10)
	defer gPool.Release()

	for msg := range e.MessageChan {
		select {
		case <-ctx.Done():
			break
		default:
			filter, err := e.filterTree.MathFilter(msg.Topic)
			if err != nil {
				log.Printf("math topic failed: %v", err)
				continue
			}

			key := msg.FromTag + "-" + filter
			tTags, ok := e.toTopicMap[key]
			if !ok {
				log.Printf("[match toTags] not match -- fTag=%v, filter=%v", msg.FromTag, filter)
				continue
			}

			for _, tTag := range tTags {
				if tTag == msg.FromTag {
					// 防止MQTT广播风暴
					continue
				}

				client, ok := e.cliConnMap[tTag]
				if !ok {
					log.Printf("[match toTags] not match client conn: key=%s", tTags)
					continue
				}

				if err = gPool.Submit(func() {
					xmqtt.FastSend(client, msg.Topic, msg.Qos, false, msg.Payload)
				}); err != nil {
					log.Printf("[submit message] failed: %v", err)
				}
			}
		}
	}
}

func (e *Engine) Start(ctx context.Context) error {
	// 接收并处理数据
	go e.handlerMessage(ctx)

	mcCap := cap(e.MessageChan)
	notifyV := int(float32(mcCap) * 0.75)

	// 开始订阅
	for tag, client := range e.cliConnMap {
		v, ok := e.cliSubMap[tag]
		if !ok {
			continue
		}
		filters := make(map[string]byte, len(v))
		for _, sub := range v {
			filters[sub.Topic] = sub.Qos
			log.Printf("[subscribe] %v :: %v", tag, sub.Topic)
		}
		_tag := tag
		client.SubscribeMultiple(filters, func(client mqtt.Client, message mqtt.Message) {
			m := Message{
				FromTag:  _tag,
				Topic:    message.Topic(),
				Payload:  message.Payload(),
				Qos:      message.Qos(),
				Retained: message.Retained(),
			}

			select {
			case <-ctx.Done():
				return
			default:
				mcSize := len(e.MessageChan)
				if mcSize > notifyV {
					log.Printf("[channel message] current channel size: %d", mcSize)
				}

				if mcSize < mcCap {
					e.MessageChan <- &m
				} else {
					log.Printf("[skip message] message channel amass; size=%d", len(e.MessageChan))
				}
			}
		})
	}

	return nil
}

func (e *Engine) Release() {
	for tag, client := range e.cliConnMap {
		v, ok := e.cliSubMap[tag]
		if !ok {
			continue
		}

		topics := make([]string, 0, len(v))
		for _, sub := range v {
			topics = append(topics, sub.Topic)
		}

		err := xmqtt.UnSubscribe(client, topics...)
		if err != nil {
			log.Printf("Unsubscribed clientTag: %v fialed %v", tag, err)
		}
	}
}

func (e *Engine) Close() {
	e.Release()
	for _, client := range e.cliConnMap {
		xmqtt.Close(client)
	}
}
