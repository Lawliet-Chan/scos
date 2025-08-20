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
	RPC          string
	ChainID      string
	SCOSAddress  string
	VaultAddress string
}

func LoadConfig() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		PrivateKey: "01c7939dc6827ee10bb7d26f420618c4af88c0029aa70be202f1ca7f29fe5bb4", // 写死私钥
		Chains: map[string]ChainInfo{
			"Reddio": {
				RPC:          "https://reddio-dev.reddio.com/",
				ChainID:      "50341",
				SCOSAddress:  "0xeB5e9Af4b798ec27A0f24DA22C7A7b3b657D05d9",
				VaultAddress: "0x0fE2B0c6177c79278A70E825581c691856E932D3",
			},
			"Scroll": {
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
