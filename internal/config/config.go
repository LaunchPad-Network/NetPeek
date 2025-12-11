package config

import (
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var initEnvOnce sync.Once
var initLocalOnce sync.Once

func init() {
	viper.SetConfigType("toml")
	initEnv()
	initLocal()
}

func initEnv() {
	initEnvOnce.Do(func() {
		viper.SetEnvPrefix("LG")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		viper.BindEnv("net.host")
		viper.BindEnv("net.port")
		viper.BindEnv("authentication.privatekey")
		viper.BindEnv("authentication.publickey")
		viper.BindEnv("log.level")
	})
}

func initLocal() {
	initLocalOnce.Do(func() {
		if p := os.Getenv("LG_CONFIG"); p != "" {
			viper.SetConfigFile(p)
			_ = viper.ReadInConfig()
			return
		}

		_, filename, _, _ := runtime.Caller(0)
		root := path.Dir(path.Dir(path.Dir(filename)))
		viper.AddConfigPath(root)
		viper.SetConfigName("config")
		err := viper.ReadInConfig()
		if err != nil {
			println("Local Log init fail")
		}
	})
}
