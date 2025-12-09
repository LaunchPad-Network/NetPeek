package config

import (
	"sync/atomic"

	"github.com/spf13/viper"
)

var isDevelopment atomic.Bool

func init() {
	isDevelopment.Store(viper.GetString("environment") == "development")
}

func IsDevelopment() bool {
	return isDevelopment.Load()
}

func Development() {
	isDevelopment.Store(true)
}

func UnDevelopment() {
	isDevelopment.Store(false)
}
