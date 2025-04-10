package main

import (
	"geo/internal/app"
	"geo/internal/config"
)

func main() {
	cfg := config.MustLoadConfig("config/local.yaml")
	app.Run(cfg)
}
