package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/alxyng/tracer/controller"
	"github.com/alxyng/tracer/internal/config"
	"github.com/alxyng/tracer/internal/modbus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

const ServiceName = "tracer-controller"

func main() {
	ctx := context.Background()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg, err := config.Get()
	if err != nil {
		logger.Fatal("error getting config", zap.Error(err))
	}

	handler, client, err := modbus.Connect(cfg.Modbus)
	if err != nil {
		logger.Fatal("unable to initiate Modbus RTU connection", zap.Error(err))
	}
	defer handler.Close()

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

	service := controller.NewService(client, logger)
	service.OnRecord(func(ctx context.Context, reading *controller.Reading) {
		payload, err := json.Marshal(reading)
		if err != nil {
			logger.Error("error marshalling reading", zap.Error(err))
			return
		}
		if token := mqttClient.Publish("tracer/reading", 0, false, payload); token.Wait() && token.Error() != nil {
			logger.Error("error publishing to mqtt", zap.String("topic", "tracer/reading"), zap.Error(token.Error()))
			return
		}
	})

	mqttTransport := controller.NewMQTTTransport(mqttClient, service, logger)
	mqttTransport.Register()

	logger.Info("running controller")
	service.Run(ctx)
}
