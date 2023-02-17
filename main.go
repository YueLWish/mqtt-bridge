package main

import (
	"context"
	"flag"
	"github.com/pkg/errors"
	"github.com/yuelwish/mqtt-bridge/engine"
	"github.com/yuelwish/mqtt-bridge/pkg/setting"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var conPath string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&conPath, "conf", "config.json", "配置文件路径 支持[josn,toml,yaml]")
	flag.Parse()

}

func Init() error {
	if err := setting.Steup(conPath); err != nil {
		return errors.Wrap(err, "配置文件存在错误")
	}
	return nil
}

func main() {
	var (
		err           error
		ctx, cancelFn = context.WithCancel(context.Background())
	)

	if err = Init(); err != nil {
		log.Fatalf("init failed: %v", err)
	}
	// 1. 初始化 all client
	// 2. 根据路由 整理出 每个 client 需要监听的 topic 和 生成 topic 匹配树
	// 3. 让 client 开始监听对应的 topic
	// 4. 收到 topic 统一处理

	eHelper := engine.NewEngineHelper()
	for _, it := range setting.AppConf.Clients {
		if err = eHelper.AddClient(it.Tag, it.Address, func(addr *engine.MqttAddress) {
			addr.UserName = it.UserName
			addr.Password = it.Password
		}); err != nil {
			log.Printf("add client failed : %v", err)
		}
	}

	for _, it := range setting.AppConf.Topics {
		if err = eHelper.AddTopicFilter(it.Tag, it.Qos, it.Filter...); err != nil {
			log.Printf("add client topic filter : %v", err)
		}
	}

	for _, it := range setting.AppConf.Routing {
		eHelper.AddRouting(it.FromTags, it.ToTags, it.TopicTags)
	}

	mEngine, err := eHelper.BuildEngine()
	if err != nil {
		log.Fatalf("BuildEngine failed : %v", err)
	}
	if err = mEngine.Dial(); err != nil {
		log.Fatalf("Dial failed : %v", err)
	}

	if err = mEngine.Start(ctx); err != nil {
		log.Fatalf("Start failed : %v", err)
	}

	// ------------- 监听杀死 -------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// --------------执行退出-----------------
	cancelFn()
	mEngine.Close()

	// ------------- 程序结束 -------------
	log.Print("app exit ...")
}
