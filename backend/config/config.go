package config

import (
	"os"
)

type Config struct {
	Port       string
	PrivateKey string
	Chains     map[string]ChainInfo
}

type ChainInfo struct {
	RPC         string
	ChainID     string
	SCOSAddress string
}

func LoadConfig() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		PrivateKey: "36a15f8d63742eaabf9ebb32a8551db13d6a3167", // 写死私钥
		Chains: map[string]ChainInfo{
			"reddio": {
				RPC:         "https://reddio-dev.reddio.com/",
				ChainID:     "50341",
				SCOSAddress: "",
			},
			"scroll": {
				RPC:         "https://sepolia-rpc.scroll.io",
				ChainID:     "534351",
				SCOSAddress: "",
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
