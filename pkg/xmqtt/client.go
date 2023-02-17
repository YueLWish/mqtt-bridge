package xmqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
	"sync/atomic"
)

type subscribed struct {
	topic    string
	qos      byte
	callback mqtt.MessageHandler
}

type Client struct {
	mqtt.Client
	subMap sync.Map
	status int32 // 1 连接成功  2 断开连接
}

const (
	stConned = iota + 1
	stLost
)

func NewClient(o *mqtt.ClientOptions) mqtt.Client {
	var (
		onLostFn mqtt.ConnectionLostHandler
		onConnFn mqtt.OnConnectHandler
	)
	if o.OnConnectionLost != nil {
		onLostFn = o.OnConnectionLost
	}
	if o.OnConnect != nil {
		onConnFn = o.OnConnect
	}

	mClient := &Client{}
	o.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		atomic.StoreInt32(&mClient.status, stLost)
		if onLostFn != nil {
			onLostFn(client, err)
		}
	})

	o.SetOnConnectHandler(func(client mqtt.Client) {
		defer atomic.StoreInt32(&mClient.status, stConned)
		if onConnFn != nil {
			onConnFn(client)
		}

		if atomic.LoadInt32(&mClient.status) == stLost {
			var n int
			mClient.subMap.Range(func(_, value interface{}) bool {
				sub := value.(*subscribed)
				mClient.Client.Subscribe(sub.topic, sub.qos, sub.callback)
				n++
				return true
			})
			mqtt.DEBUG.Printf("[CONN resubscribe %d topic]", n)
		}
	})
	mClient.Client = mqtt.NewClient(o)

	return mClient
}

func (c *Client) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	token := c.Client.SubscribeMultiple(filters, callback)

	if token.Error() == nil {
		for topic, qos := range filters {
			sub := subscribed{
				topic:    topic,
				qos:      qos,
				callback: callback,
			}
			c.subMap.Store(sub.topic, &sub)
		}
	}
	return token
}
func (c *Client) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	token := c.Client.Subscribe(topic, qos, callback)

	if token.Error() == nil {
		sub := subscribed{
			topic:    topic,
			qos:      qos,
			callback: callback,
		}
		c.subMap.Store(sub.topic, &sub)
	}
	return token
}

func (c *Client) Unsubscribe(topics ...string) mqtt.Token {
	token := c.Client.Unsubscribe(topics...)

	if token.Error() == nil {
		for _, topic := range topics {
			c.subMap.Delete(topic)
		}
	}
	return token
}

func (c *Client) AddRoute(topic string, callback mqtt.MessageHandler) {
	c.Client.AddRoute(topic, callback)

	c.subMap.Store(topic, &subscribed{
		topic:    topic,
		qos:      0,
		callback: callback,
	})
}
func (c *Client) Disconnect(quiesce uint) {
	c.Client.Disconnect(quiesce)

	c.subMap.Range(func(key, _ interface{}) bool {
		c.subMap.Delete(key)
		return true
	})
}
