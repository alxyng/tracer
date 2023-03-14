package main

import (
	"log"
	"time"

	"github.com/alxyng/tracer/api"
	"github.com/alxyng/tracer/internal/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ServiceName = "tracer-api"

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
		AddBroker(cfg.MQTT.Broker).
		SetClientID(ServiceName).
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("error connecting to mqtt", zap.Error(token.Error()))
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	transport := api.NewHTTPTransport(router, mqttClient, logger)
	transport.Register()

	logger.Info("starting http server", zap.String("addr", cfg.API.Addr))
	router.Run(cfg.API.Addr)
}
