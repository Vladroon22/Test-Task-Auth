package config

import "os"

type Config struct {
	Addr_PORT string
	DB        string
	JWT       string
	Email     string
	AppPass   string
}

func CreateConfig() *Config {
	return &Config{
		Addr_PORT: getEnv("addr_port", ":3000"),
		DB:        getEnv("DB", ""),
		JWT:       getEnv("JWT", ""),
		Email:     getEnv("email", ""),
		AppPass:   getEnv("AppPass", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
