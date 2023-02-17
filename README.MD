# MQTT-BRIDGE

## 背景

在项目开发中，我们的设备都在生产环境中，测试环境数据只有拉拉稀稀几条数据，于是我就想把生产的服务的相关数据转发到测试环境中，MQTT桥就这样产生了。

## 使用

```bash
Usage of mqtt-bridge:
  -conf string
        配置文件路径 支持[josn,toml,yaml] (default "config.json")
```

参数就一个，配置文件默认是 json，其实我更喜欢 toml. 后面所有内容也都是以 json 为例子。



## 配置文件

配置文件主要分 3 块，`clients`、`topics` 和 `routing` 。

- clients 
- topics 
- routing 路由，即订阅到什么 topic 然后转发给 哪个客户端的描述

```json
{
  "clients": [], // 客户端配置
  "topics": [], // topic配置
  "routing": [] // 路由配置
}
```



**公共参数**

- tag： 表示在该模块下唯一的名字标识



**clients **

>  客户端信息配置

```json
{
  "tag": "test-emqx",  // 必填
  "address": "tcp://xxx.xxx.xxx:1883", // 必填
  "userName": "test", // 选填
  "password": "123456"  // 选填
}

```

- address：MQTT 的地址，支持ws、wss、tpc、udp
- userName：MQTT 的用户名，匿名登陆可以不填
- password：MQTT 的密码，匿名登陆可以不填

**topics**

> 主题配置



```json
{
  "tag": "gps-out",
  "qos": 0,
  "filter": [
    "ht1/gps/#",
    "ht2/gps/#"
  ]
}
```



- qos：MQTT 中质量保证
- filter：过滤器，其实就是订阅哪些 topic , 这是数组，可以一个 tag 订阅多个 topic



**routing**

> 路由规则配置

```json
{
  "fromTags": [
    "test-emqx"
  ],
  "toTags": [
    "gps-out"
  ],
  "topicTags": [
    "rel-emqx"
  ]
}
```

- fromTags：从哪里订阅 ，这里的 tag 指 client 的tag
- toTags：发送到哪里去，这里的 tag 指 client 的tag

- topicTags：要订阅的topic的 tag, 就是 topic 的tag

## 例子

```json
{
    "clients": [
        {
            "tag": "hivemq-pub",
            "address": "tcp://broker.hivemq.com:1883"
        },
        {
            "tag": "emqx-pub",
            "address": "tcp://broker-cn.emqx.io:1883"
        }
    ],
    "topics": [
        {
            "tag": "hh",
            "qos": 0,
            "filter": ["hh/out/#"]
        },
        {
            "tag": "ee",
            "qos": 0,
            "filter": ["ee/out/#"]
        }
    ],
    "routing": [
        {
            "fromTags": ["hivemq-pub"],
            "toTags": ["emqx-pub"],
            "topicTags": ["hh"]
        },
        {
            "fromTags": ["emqx-pub"],
            "toTags": ["hivemq-pub"],
            "topicTags": ["ee"]
        }
    ]
}
```

**解释**

这里有两个 client , 一个是 emqx-pub 另外一个是 hivemq-pub .

两个 topic , hh 和 ee

路由规则很简单：

1. 从 emqx-pub 订阅 `hh/out/#` 然后转发给 hivemq-pub
2. 从 hivemq-pub 订阅 `ee/out/#` 转发给 emqx-pub
