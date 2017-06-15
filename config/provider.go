package config

import "github.com/NYTimes/video-captions-api/providers"

// ProviderConfig is an interface for injecting CaptionsServiceConfig values into Providers
type ProviderConfig interface {
	// NewProvider creates a Provider with CaptionsServiceConfig
	NewProvider(*CaptionsServiceConfig) providers.Provider
}
