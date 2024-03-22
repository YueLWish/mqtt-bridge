package engine

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
	"github.com/yuelwish/mqtt-bridge/pkg/kit"
	"github.com/yuelwish/mqtt-bridge/pkg/logger"
	"github.com/yuelwish/mqtt-bridge/pkg/xmqtt"
	"go.uber.org/zap"
	"runtime"
)

const (
	clientId = "mqtt-bridge"
)

type Engine struct {
	cliAddrMap         map[string]*MqttAddress // clientTag : client address
	cliSubMap          map[string][]*SubTopic  // clientTag : 对应订阅的 topicFilter
	filterTree         *TopicFilterTree        // filter 按 / 分割生成树形结构
	toTopicMap         map[string][]string     // fromTag+topicFilter : toTags
	routingFilterTable map[string]string       // 路由中的 filter 映射 tag

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
		sCtxt := msg.ctx
		select {
		case <-ctx.Done():
			break
		default:
			filter, err := e.filterTree.MathFilter(msg.Topic)
			if err != nil {
				logger.WithContext(sCtxt).Info("math topic failed", zap.Error(err))
				continue
			}

			key := msg.FromTag + "-" + filter
			tTags, ok := e.toTopicMap[key]
			if !ok {
				logger.WithContext(ctx).Info("[match toTags] not match", zap.String("fTag", msg.FromTag), zap.String("filter", filter))
				continue
			}

			for _, tTag := range tTags {
				if tTag == msg.FromTag {
					// 防止MQTT广播风暴
					continue
				}

				client, ok := e.cliConnMap[tTag]
				if !ok {
					logger.WithContext(ctx).Info("[match toTags] not match client conn", zap.String("tTag", tTag))
					continue
				}

				if err = gPool.Submit(func() {
					logger.WithContext(ctx).Debug("relay the message",
						zap.String("routTag", e.routingFilterTable[filter]),
						zap.String("fromTag", msg.FromTag),
						zap.String("toTag", tTag),
						zap.String("topic", msg.Topic),
						zap.ByteString("payload", msg.Payload),
					)
					xmqtt.MustPublish(client, msg.Topic, msg.Qos, msg.Retained, msg.Payload)
				}); err != nil {
					logger.WithContext(ctx).Error("[submit message] failed", zap.Error(err))
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
			logger.Info("[subscribe]", zap.String("clientTag", tag), zap.String("topic", sub.Topic))
		}
		_tag := tag
		client.SubscribeMultiple(filters, func(client mqtt.Client, message mqtt.Message) {
			sCtx, cancelFunc := context.WithCancel(ctx)
			defer cancelFunc()
			sCtx = e.newCtx(sCtx)

			m := Message{
				ctx:      sCtx,
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
					logger.WithContext(sCtx).Info("[channel message] current channel", zap.Int("size", mcSize))
				}

				if mcSize < mcCap {
					e.MessageChan <- &m
				} else {
					logger.WithContext(sCtx).Info("[skip message] message channel amass", zap.Int("size", len(e.MessageChan)))
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
			logger.Warn("Unsubscribed clientTag", zap.String("clientTag", tag), zap.Error(err))
		}
	}
}

func (e *Engine) newCtx(ctx context.Context) context.Context {
	return logger.NewContext(ctx, zap.String("traceId", kit.NewTraceId()))
}

func (e *Engine) Close() {
	e.Release()
	for _, client := range e.cliConnMap {
		xmqtt.Close(client)
	}
}
