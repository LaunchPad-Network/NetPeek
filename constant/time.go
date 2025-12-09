package constant

import "time"

const (
	TimeFormat = "2006-01-02 15:04:05"

	ProxyReqSignValidityDuration = 30 * time.Second

	ServersListPullInterval     = 10 * time.Minute
	ServersListMinPullInterval  = 1 * time.Minute
	BGPCommunityDefPullInterval = 10 * time.Minute
)
