package serverslist

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LaunchPad-Network/NetPeek/constant"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/net"

	"github.com/lfcypo/viperx"
	"github.com/spf13/viper"
)

var servers *[]*Server
var serversLock sync.RWMutex
var startPullOnce sync.Once
var noticeFetchCh = make(chan struct{}, 1)
var lastPullTime time.Time
var lastPullTimeLock sync.Mutex

func GetServersList() []*Server {
	serversLock.RLock()
	defer serversLock.RUnlock()

	if servers == nil {
		return []*Server{}
	}

	return *servers
}

func GetServerByID(id string) *Server {
	srvList := GetServersList()
	for _, srv := range srvList {
		if strings.EqualFold(srv.Id, id) {
			return srv
		}
	}
	return nil
}

func fetchServersList(url string) ([]*Server, error) {
	r, err := net.FetchURLWithTimeoutAsPlaintext(url, viperx.GetInt("servers.timeout", 5))
	if err != nil {
		return nil, err
	}

	srvList, err := ParseCSV(r)
	if err != nil {
		return nil, err
	}

	return srvList, nil
}

func pullServersList() {
	url := viper.GetString("servers.pull_url")
	if url == "" {
		log.Fatal("servers.pull_url is empty, need it to pull PoP list")
	}

	srvList, err := fetchServersList(url + "?t=" + strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		log.Errorf("failed to pull servers list: %v", err)
		return
	}

	serversLock.Lock()
	servers = &srvList
	serversLock.Unlock()

	lastPullTimeLock.Lock()
	lastPullTime = time.Now()
	lastPullTimeLock.Unlock()

	log.Infof("successfully pulled %d servers from %s", len(srvList), url)
}

func StartPullingServersList(stopCh <-chan struct{}) {
	startPullOnce.Do(func() {
		go func() {
			log.Info("starting servers list pulling...")

			pullServersList()

			ticker := time.NewTicker(constant.ServersListPullInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
				case <-noticeFetchCh:
					if time.Since(lastPullTime) < constant.ServersListMinPullInterval {
						log.Debug("skipping servers list pulling due to min pull interval")
					} else {
						log.Debug("pulling servers list...")
						pullServersList()
					}
				case <-stopCh:
					log.Info("stopping servers list pulling")
					return
				}
			}
		}()
	})
}

func NotifyFetchServersList() {
	select {
	case noticeFetchCh <- struct{}{}:
	default:
	}
}
