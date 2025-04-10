package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type Config struct {
	Dadata     `yaml:"dadata"`
	Geoservice `yaml:"geoservice"`
	Token      `yaml:"token"`
}

type Dadata struct {
	ApiKey    string `yaml:"api_key" env:"API_KEY"`
	ApiSecret string `yaml:"api_secret" env:"API_SECRET"`
}

type Geoservice struct {
	Host            string        `yaml:"host" env:"GEOSERVICE_HOST" env-default:""`
	Port            string        `yaml:"port" env:"GEOSERVICE_PORT" env-default:":8080"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"30s"`
}

type Token struct {
	Secret string        `yaml:"secret" env:"TOKEN_SECRET"`
	TTL    time.Duration `yaml:"ttl" env:"TTL" env-default:"10m"`
	Skew   time.Duration `yaml:"skew" env:"TOKEN_SKEW" env-default:"30s"`
}

func MustLoadConfig(path string) *Config {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		log.Fatal("cannot read config: ", err)
	}
	return cfg
}
