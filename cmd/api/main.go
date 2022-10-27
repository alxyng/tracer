package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alxyng/tracer/api"
	"github.com/alxyng/tracer/internal/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg, err := config.Get()
	if err != nil {
		logger.Fatal("error getting config", zap.Error(err))
	}

	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("tracer-api").
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("error connecting to mqtt", zap.Error(token.Error()))
	}

	router := mux.NewRouter()
	transport := api.NewHTTPTransport(router, mqttClient, logger)
	transport.Register()

	server := http.Server{Addr: ":" + cfg.API.Port, Handler: router}
	logger.Info("starting api server", zap.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("error running server", zap.Error(err))
	}
}
