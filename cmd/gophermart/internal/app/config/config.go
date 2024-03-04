package config

import (
	"flag"
	"os"
)

// - адрес и порт запуска сервиса: переменная окружения ОС `RUN_ADDRESS` или флаг `-a`
// - адрес подключения к базе данных: переменная окружения ОС `DATABASE_URI` или флаг `-d`
// - адрес системы расчёта начислений: переменная окружения ОС `ACCRUAL_SYSTEM_ADDRESS` или флаг `-r`

type Config struct {
	RunAddress          string
	DatabaseURI         string
	AcrualSystemAddress string
}

func New() *Config {
	config := &Config{}
	config.setByEnvs()
	config.setByFlags()
	config.setDefault()
	return config
}

func (c *Config) setDefault() {
	if c.RunAddress == "" {
		c.RunAddress = "localhost:8080"
	}
	if c.DatabaseURI == "" {
		c.DatabaseURI = "postgresql://user:user@localhost:5432/gophermart?sslmode=disable"
	}
	if c.AcrualSystemAddress == "" {
		c.AcrualSystemAddress = "localhost:8081"
	}
}

func (c *Config) setByEnvs() {
	if c.RunAddress == "" {
		c.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if c.DatabaseURI == "" {
		c.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if c.AcrualSystemAddress == "" {
		c.AcrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}
}

func (c *Config) setByFlags() {
	var runAddress string
	flag.StringVar(&runAddress, "a", "localhost:8080", "run address")
	c.RunAddress = runAddress
	var databaseURI string
	flag.StringVar(&databaseURI, "d", "postgresql://user:user@localhost:5432/gophermart?sslmode=disable", "database URI")
	c.DatabaseURI = databaseURI
	var acrualSystemAddress string
	flag.StringVar(&acrualSystemAddress, "r", "localhost:8081", "accrual system address")
	c.AcrualSystemAddress = acrualSystemAddress
}
