package service

import (
	"net/http"
	"time"

	"github.com/jungtechou/valomap/config"
	domain "github.com/jungtechou/valomap/domain/map"
	"github.com/jungtechou/valomap/service/cache"
	"github.com/jungtechou/valomap/service/roulette"

	"github.com/google/wire"
)

// ProvideMapPool creates and returns a MapPool with predefined maps
func ProvideMapPool() domain.MapPool {
	return domain.MapPool{
		Maps: []domain.Map{
			{UUID: "7eaecc1b-4337-bbf6-6ab9-04b8f06b3319", DisplayName: "Ascent"},
			{UUID: "d960549e-485c-e861-8d71-aa9d1aed12a2", DisplayName: "Split"},
			{UUID: "b529448b-4d60-346e-e89e-00a4c527a405", DisplayName: "Fracture"},
			{UUID: "2c9d57ec-4431-9c5e-2939-8f9ef6dd5cba", DisplayName: "Bind"},
			{UUID: "2fb9a4fd-47b8-4e7d-a969-74b4046ebd53", DisplayName: "Breeze"},
			{UUID: "e2ad5c54-4114-a870-9641-8ea21279579a", DisplayName: "Icebox"},
			{UUID: "ee613ee9-28b7-4beb-9666-08db13bb2244", DisplayName: "Haven"},
			{UUID: "fd267378-4d1d-484f-ff52-77821ed10dc2", DisplayName: "Pearl"},
			{UUID: "92584fbe-486a-b1b2-28f9-d29d88a7c7ca", DisplayName: "Lotus"},
			{UUID: "690b3152-4c7a-d5a2-0093-9796b54176ce", DisplayName: "Sunset"},
		},
	}
}

// ProvideHTTPClient creates and returns an HTTP client with reasonable defaults
func ProvideHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// ProvideImageCache creates and returns an image cache service
func ProvideImageCache(cfg *config.Config, client *http.Client) (cache.ImageCache, error) {
	return cache.NewImageCache(cfg, client)
}

var ServiceSet = wire.NewSet(
	roulette.NewService,
	ProvideMapPool,
	ProvideHTTPClient,
	ProvideImageCache,
)
