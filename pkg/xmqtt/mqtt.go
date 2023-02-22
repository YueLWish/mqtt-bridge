package xmqtt

import (
	"log"
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
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetMaxReconnectInterval(10 * time.Second)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		r := client.OptionsReader()
		cId := r.ClientID()

		urls := r.Servers()
		hosts := make([]string, len(urls))
		for i, url := range urls {
			hosts[i] = url.Host
		}
		hostLink := strings.Join(hosts, ", ")

		log.Printf("[CONN successful] -- %s conn %s", cId, hostLink)
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

		log.Printf("[CONN ERROR] -- %s conn %s %v", cId, hostLink, err)
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

func UnSubscribe(client mqtt.Client, topic ...string) error {
	token := client.Unsubscribe(topic...)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func Send(client mqtt.Client, topic string, qos byte, retained bool, payload []byte) error {
	token := client.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func Subscribe(client mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) error {
	token := client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
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
