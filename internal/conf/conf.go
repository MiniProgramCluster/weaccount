package conf

import (
	"weaccount/utils/log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type DatabaseConf struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Database    string `json:"database"`
	PoolMaxOpen int    `json:"poolMaxOpen"`
	PoolMaxIdle int    `json:"poolMaxIdle"`
}

type AppConf struct {
	AppID     string `json:"appid"`
	AppSecret string `json:"appSecret"`
}

type TokenConf struct {
	LifeTime int64  `json:"life_time"`
	Secret   string `json:"secret"`
}

type config struct {
	Apps     map[string]*AppConf `json:"apps"`
	Token    TokenConf
	Database DatabaseConf
}

var conf config

func parseAppConfig() {
	if err := viper.Unmarshal(&conf); err != nil {
		log.Logger().Error().Err(err).Msg("Failed to unmarshal app config")
	}
	for appID, appConf := range conf.Apps {
		appConf.AppID = appID
	}
}

func Init(file string) {
	viper.OnConfigChange(func(e fsnotify.Event) {
		conf.Apps = nil
		parseAppConfig()
	})
	// 使用 viper 加载 JSON 配置文件
	viper.SetConfigFile(file)   // 配置文件路径
	viper.SetConfigType("json") // 配置文件类型
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Logger().Fatal().Err(err).Msg("Error reading config file")
	}
	parseAppConfig()
}

func App(appID string) *AppConf {
	return conf.Apps[appID]
}

func Token() *TokenConf {
	return &conf.Token
}

func Database() *DatabaseConf {
	return &conf.Database
}
