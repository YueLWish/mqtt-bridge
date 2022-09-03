package setting

import (
	"github.com/spf13/viper"
	"log"
)

type AppConfig struct {
	Clients []ClientConfig  `json:"clients"`
	Topics  []TopicConfig   `json:"topics"`
	Routing []RoutingConfig `json:"routing"`
}
type ClientConfig struct {
	Tag      string `json:"tag"`
	Address  string `json:"address"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}
type TopicConfig struct {
	Tag    string   `json:"tag"`
	Qos    byte     `json:"qos"`
	Filter []string `json:"filter"`
}
type RoutingConfig struct {
	FromTags  []string `json:"fromTags"`
	ToTags    []string `json:"toTags"`
	TopicTags []string `json:"topicTags"`
}

var AppConf = &AppConfig{}

func Steup(configFile string) error {
	viper.SetConfigFile(configFile)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("viper.ReadInConfig() failed: %v", err)
		return err
	}
	err = viper.Unmarshal(AppConf)
	if err != nil {
		log.Printf("viper.Unmarshal() failed: %v", err)
		return err
	}
	return nil
}
