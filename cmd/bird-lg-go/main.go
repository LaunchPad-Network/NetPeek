package main

import (
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/banner"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/communityparser"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/serverslist"
	"github.com/LaunchPad-Network/NetPeek/internal/service/frontend"

	"github.com/lfcypo/viperx"
)

var log = logger.New("Main")

func main() {
	banner.PrintBanner("")

	stopChan := make(chan struct{})
	defer close(stopChan)

	serverslist.StartPullingServersList(stopChan)
	communityparser.StartPulling(stopChan)

	r := frontend.SetupRouter()

	log.Info("Listening on " + viperx.GetString("net.host", "0.0.0.0") + ":" + viperx.GetString("net.port", "1790"))
	err := r.Run(viperx.GetString("net.host", "0.0.0.0") + ":" + viperx.GetString("net.port", "1790"))

	if err != nil {
		log.Fatal(err)
	}

	select {
	case stopChan <- struct{}{}:
	default:
	}
}
