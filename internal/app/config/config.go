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
	IsMigrate           bool
}

func New() *Config {
	config := &Config{}
	config.setByFlags()
	config.setByEnvs()
	return config
}

func (c *Config) setByEnvs() {
	if val := os.Getenv("RUN_ADDRESS"); val != "" {
		c.RunAddress = val
	}
	if val := os.Getenv("DATABASE_URI"); val != "" {
		c.DatabaseURI = val
	}
	if val := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); val != "" {
		c.AcrualSystemAddress = val
	}
}

func (c *Config) setByFlags() {
	var runAddress string
	flag.StringVar(&runAddress, "a", "localhost:8081", "run address")
	c.RunAddress = runAddress
	var databaseURI string
	flag.StringVar(&databaseURI, "d", "postgresql://user:user@localhost:5432/gophermart", "database URI")
	c.DatabaseURI = databaseURI
	var acrualSystemAddress string
	flag.StringVar(&acrualSystemAddress, "r", "http://localhost:8080", "accrual system address")
	c.AcrualSystemAddress = acrualSystemAddress
	var isMigrate bool
	flag.BoolVar(&isMigrate, "m", false, "is migrate")
	c.IsMigrate = isMigrate
}
