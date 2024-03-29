package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alxyng/tracer/controller"
	"github.com/alxyng/tracer/internal/config"
	"github.com/alxyng/tracer/writer"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const ServiceName = "tracer-writer"

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

	conn, err := pgx.Connect(ctx, cfg.Database.DSN)
	if err != nil {
		logger.Fatal("unable to connect to database", zap.Error(err))
	}
	defer conn.Close(ctx)
	logger.Info("connected to database", zap.String("database", conn.Config().Database))

	w1 := writer.NewSQLWriter(conn, logger)
	w2 := writer.NewSQLAggregateWriter(conn, logger)
	var numConns uint64
	var numConnLosts uint64

	onConnect := func(c mqtt.Client) {
		numConns++
		logger.Info("connected to mqtt")

		handler := func(client mqtt.Client, msg mqtt.Message) {
			var reading controller.Reading
			if err := json.Unmarshal(msg.Payload(), &reading); err != nil {
				logger.Error("error unmarshalling reading", zap.Error(err))
				return
			}

			if err := w1.Write(ctx, reading); err != nil {
				logger.Fatal("error writing", zap.String("writer", "w1"), zap.Error(err))
			}

			if err := w2.Write(ctx, reading); err != nil {
				logger.Fatal("error writing", zap.String("writer", "w2"), zap.Error(err))
			}
		}

		if token := c.Subscribe("tracer/reading", 0, handler); token.Wait() && token.Error() != nil {
			logger.Fatal("error subscribing to mqtt topic", zap.String("topic", "tracer/reading"), zap.Error(token.Error()))
		}
	}

	onConnectionLost := func(c mqtt.Client, err error) {
		numConnLosts++
		logger.Error("mqtt connection lost", zap.Error(err))
	}

	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTT.Broker).
		SetClientID(ServiceName).
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second).
		SetOnConnectHandler(onConnect).
		SetConnectionLostHandler(onConnectionLost)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatal("error connecting to mqtt", zap.Error(token.Error()))
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	router.GET("/", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "w1Writes": %v,
  "w2Writes": %v,
  "mqttConnected": %v,
  "mqttConnectionOpen": %v,
  "numConns": %v,
  "numConnLosts": %v
}`, w1.NumWrites(), w2.NumWrites(), mqttClient.IsConnected(), mqttClient.IsConnectionOpen(), numConns, numConnLosts)
	}))
	logger.Info("starting http server", zap.String("addr", ":3011"))
	router.Run(":3011")
}
