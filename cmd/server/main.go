package main

import (
	"cm_water_tg/internal/bot"
	"cm_water_tg/internal/config"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()

	go bot.StartBot(cfg)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
