package config

import "github.com/spf13/viper"

type BrandingInfo struct {
	Name     string
	MapUrl   string
	LgDomain string
}

func GetBrandingInfo() BrandingInfo {
	branding := BrandingInfo{
		Name:     viper.GetString("branding.name"),
		MapUrl:   viper.GetString("branding.map_url"),
		LgDomain: viper.GetString("branding.lg_domain"),
	}

	if branding.Name == "" {
		branding.Name = "NetPeek"
	}
	if branding.LgDomain == "" {
		branding.LgDomain = "netpeek.local"
	}

	return branding
}
