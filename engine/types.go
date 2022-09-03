package engine

type MqttAddress struct {
	Address  string
	UserName string
	Password string
}

type TopicFilter struct {
	Tag    string
	Qos    byte
	Filter []string
}

type SubTopic struct {
	Topic string
	Qos   byte
}

type Router struct {
	FromTags  []string
	ToTags    []string
	TopicTags []string
}

type Node struct {
	Value string
	Child map[string]*Node
}

type Message struct {
	FromTag string
	Topic   string
	Payload []byte
}
