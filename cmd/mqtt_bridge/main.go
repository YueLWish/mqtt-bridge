package main

import (
	"context"
	"flag"
	"github.com/yuelwish/mqtt-bridge/pkg/errors"

	"github.com/yuelwish/mqtt-bridge/engine"
	"github.com/yuelwish/mqtt-bridge/pkg/logger"
	"github.com/yuelwish/mqtt-bridge/pkg/setting"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var conPath string

func init() {
	flag.StringVar(&conPath, "conf", "config.json", "配置文件路径 支持[josn,toml,yaml]")
	flag.Parse()
}

func Init() error {
	if err := setting.Steup(conPath); err != nil {
		return errors.WithMessage(err, "配置文件存在错误")
	}
	return nil
}

func main() {
	var (
		err           error
		ctx, cancelFn = context.WithCancel(context.Background())
	)

	if err = Init(); err != nil {
		logger.Fatal("init failed", zap.Error(err))
	}

	_, syncFn, err := logger.NewLogger(setting.AppConf)
	if err != nil {
		logger.Fatal("new logger failed", zap.Error(err))
	}
	defer syncFn()

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
			logger.Warn("add client failed", zap.Error(err))
		}
	}

	for _, it := range setting.AppConf.Topics {
		if err = eHelper.AddTopicFilter(it.Tag, it.Qos, it.Filter...); err != nil {
			logger.Warn("add client topic filter failed ", zap.Error(err))
		}
	}

	for _, it := range setting.AppConf.Routing {
		eHelper.AddRouting(it.FromTags, it.ToTags, it.TopicTags)
	}

	mEngine, err := eHelper.BuildEngine()
	if err != nil {
		logger.Fatal("BuildEngine failed", zap.Error(err))
	}
	if err = mEngine.Dial(); err != nil {
		logger.Fatal("Dial failed", zap.Error(err))
	}

	if err = mEngine.Start(ctx); err != nil {
		logger.Fatal("Start failed", zap.Error(err))
	}

	// ------------- 监听杀死 -------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// --------------执行退出-----------------
	mEngine.Close()
	cancelFn()

	// ------------- 程序结束 -------------
	logger.Info("app exit ...")
}
