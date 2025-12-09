package communityparser

import (
	"sync"
	"time"

	"github.com/LaunchPad-Network/NetPeek/constant"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/net"

	"github.com/lfcypo/viperx"
	"github.com/spf13/viper"
)

var processors []BGPCommunityProcessor
var processorsLock sync.RWMutex
var configOnce sync.Once
var startPullOnce sync.Once
var configErr error
var configf *config

func init() {
	loadConfig()
}

func ProcessBirdOutput(output string) string {
	processorsLock.RLock()
	defer processorsLock.RUnlock()

	for i := range processors {
		output = processors[i].FormatBGPText(output)
	}

	return output
}

type communityListEntry struct {
	Prefix string `mapstructure:"prefix"`
	Url    string `mapstructure:"url"`
}

type config struct {
	BGPCommunities struct {
		List []communityListEntry `mapstructure:"list"`
	} `mapstructure:"bgp_communities"`
}

func loadConfig() (*config, error) {
	configOnce.Do(func() {
		var cfg config
		configErr = viper.Unmarshal(&cfg)
		if configErr == nil {
			configf = &cfg
		}
	})
	return configf, configErr
}

func (ent communityListEntry) Fetch() (string, error) {
	r, err := net.FetchURLWithTimeoutAsPlaintext(ent.Url, viperx.GetInt("servers.timeout", 5))
	if err != nil {
		return "", err
	}
	return r, nil
}

func pullCommunityList() {
	processorsLock.Lock()
	defer processorsLock.Unlock()

	processors = nil
	processors = append(processors, *NewBGPCommunityProcessor(rfcdefs, ""))

	for _, entry := range configf.BGPCommunities.List {
		ret, err := entry.Fetch()
		if err != nil {
			log.Errorf("fetch bgp community list fail, url %s", entry.Url)
		} else {
			log.Infof("bgp community list fetch ok, url %s, prefix %s", entry.Url, entry.Prefix)
			processors = append(processors, *NewBGPCommunityProcessor(ret, entry.Prefix))
		}
	}
}

func StartPulling(stopCh <-chan struct{}) {
	startPullOnce.Do(func() {
		go func() {
			log.Info("starting community def pulling...")

			pullCommunityList()

			ticker := time.NewTicker(constant.BGPCommunityDefPullInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					pullCommunityList()
				case <-stopCh:
					log.Info("stopping servers list pulling")
					return
				}
			}
		}()
	})
}
