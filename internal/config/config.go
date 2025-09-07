package config

import "os"

type Config struct {
	DbHost     string
	DbUser     string
	DbPassword string
	DbPort     string
	DbName     string
}

func LoadConfig() *Config {
	return &Config{
		DbHost:     os.Getenv("DB_HOST"),
		DbUser:     os.Getenv("DB_USER"),
		DbPassword: os.Getenv("DB_PASSWORD"),
		DbPort:     os.Getenv("DB_PORT"),
		DbName:     os.Getenv("DB_NAME"),
	}
}
