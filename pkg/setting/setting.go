package setting

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Clients []ClientConfig  `mapstructure:"clients"`
	Topics  []TopicConfig   `mapstructure:"topics"`
	Routing []RoutingConfig `mapstructure:"routing"`
	Log     Log             `mapstructure:"log"`
}
type ClientConfig struct {
	Tag      string `mapstructure:"tag"`
	Address  string `mapstructure:"address"`
	UserName string `mapstructure:"userName"`
	Password string `mapstructure:"password"`
}
type TopicConfig struct {
	Tag    string   `mapstructure:"tag"`
	Qos    byte     `mapstructure:"qos"`
	Filter []string `mapstructure:"filter"`
}
type RoutingConfig struct {
	FromTags  []string `mapstructure:"fromTags"`
	ToTags    []string `mapstructure:"toTags"`
	TopicTags []string `mapstructure:"topicTags"`
}

type Log struct {
	Level      string `mapstructure:"level"`
	Console    bool   `mapstructure:"console"`
	Encoder    string `mapstructure:"encoder"`
	FileName   string `mapstructure:"fileName"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxAge     int    `mapstructure:"maxAge"`
	MaxBackups int    `mapstructure:"maxBackups"`
}

var AppConf = &AppConfig{}

func Steup(configFile string) error {
	viper.SetConfigFile(configFile)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrapf(err, "viper.ReadInConfig() failed: %v", err)
	}
	err = viper.Unmarshal(AppConf)
	if err != nil {
		return errors.Wrapf(err, "viper.Unmarshal() failed: %v", err)
	}
	return nil
}
