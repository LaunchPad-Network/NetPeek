package main

import (
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/banner"
	"github.com/LaunchPad-Network/NetPeek/internal/service/proxy"

	"github.com/lfcypo/viperx"
)

var log = logger.New("Main")

func main() {
	banner.PrintBanner("Proxy")

	r := proxy.SetupRouter()

	log.Info("Listening on " + viperx.GetString("net.host", "0.0.0.0") + ":" + viperx.GetString("net.port", "10179"))
	err := r.Run(viperx.GetString("net.host", "0.0.0.0") + ":" + viperx.GetString("net.port", "10179"))

	if err != nil {
		log.Fatal(err)
	}
}
