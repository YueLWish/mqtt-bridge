package xmqtt

import (
	"github.com/yuelwish/mqtt-bridge/pkg/logger"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	timeoutDuration = time.Second * 7
)

func Init(clientIdPrefix, addr string, optFns ...func(opt *mqtt.ClientOptions)) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(addr)
	opts.SetClientID(clientIdPrefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 36))

	opts.ConnectRetry = true
	opts.ConnectTimeout = 10 * time.Second
	opts.WriteTimeout = 15 * time.Second
	opts.ResumeSubs = true
	opts.Order = false

	opts.KeepAlive = 15
	opts.PingTimeout = 60 * time.Second
	opts.MaxReconnectInterval = 30 * time.Second

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		r := client.OptionsReader()
		cId := r.ClientID()

		urls := r.Servers()
		hosts := make([]string, len(urls))
		for i, url := range urls {
			hosts[i] = url.Host
		}
		hostLink := strings.Join(hosts, ", ")

		logger.Debug("[CONN successful]", zap.String("client_id", cId), zap.String("host_link", hostLink))
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		r := client.OptionsReader()
		cId := r.ClientID()

		urls := r.Servers()
		hosts := make([]string, len(urls))
		for i, url := range urls {
			hosts[i] = url.Host
		}
		hostLink := strings.Join(hosts, ", ")
		logger.Debug("[CONN ERROR]", zap.String("client_id", cId), zap.String("host_link", hostLink), zap.Error(err))
	})

	for _, fn := range optFns {
		fn(opts)
	}

	client := NewClient(opts)

	if token := client.Connect(); token.WaitTimeout(timeoutDuration) && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func MustUnSubscribe(client mqtt.Client, topic ...string) {
	_ = client.Unsubscribe(topic...)
}
func UnSubscribe(client mqtt.Client, topic ...string) error {
	token := client.Unsubscribe(topic...)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func MustPublish(client mqtt.Client, topic string, qos byte, retained bool, payload []byte) {
	_ = client.Publish(topic, qos, retained, payload)
}

func Publish(client mqtt.Client, topic string, qos byte, retained bool, payload []byte) error {
	token := client.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func MustSubscribe(client mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) {
	_ = client.Subscribe(topic, qos, callback)
}
func Subscribe(client mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) error {
	token := client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func MustSubscribes(client mqtt.Client, topics []string, qos byte, callback mqtt.MessageHandler) {
	filters := make(map[string]byte, len(topics))
	for _, topic := range topics {
		filters[topic] = qos // topic:qos
	}

	_ = client.SubscribeMultiple(filters, callback)
}
func Subscribes(client mqtt.Client, topics []string, qos byte, callback mqtt.MessageHandler) error {
	filters := make(map[string]byte, len(topics))
	for _, topic := range topics {
		filters[topic] = qos // topic:qos
	}

	token := client.SubscribeMultiple(filters, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func Close(client mqtt.Client) {
	client.Disconnect(250)
}
