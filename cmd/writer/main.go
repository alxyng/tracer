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
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

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

	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("tracer-writer").
		SetKeepAlive(2 * time.Second).
		SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatal("error connecting to mqtt", zap.Error(token.Error()))
	}

	w1 := writer.NewSQLWriter(conn, logger)
	w2 := writer.NewSQLAggregateWriter(conn, logger)

	handler := func(client mqtt.Client, msg mqtt.Message) {
		var reading controller.Reading
		if err := json.Unmarshal(msg.Payload(), &reading); err != nil {
			logger.Error("error unmarshalling reading", zap.Error(err))
			return
		}

		if err := w1.Write(ctx, reading); err != nil {
			logger.Error("error writing", zap.String("writer", "w1"), zap.Error(err))
		}

		if err := w2.Write(ctx, reading); err != nil {
			logger.Error("error writing", zap.String("writer", "w2"), zap.Error(err))
		}
	}

	if token := mqttClient.Subscribe("tracer/reading", 0, handler); token.Wait() && token.Error() != nil {
		logger.Fatal("error subscribing to mqtt topic", zap.String("topic", "tracer/reading"), zap.Error(token.Error()))
	}

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pgOk := true
		if err := conn.Ping(r.Context()); err != nil {
			logger.Error("error pinging database", zap.Error(err))
			pgOk = false
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "w1Writes": %v,
  "w2Writes": %v,
  "mqttConnected": %v,
  "mqttConnectionOpen": %v,
  "pgOk": %v
}`, w1.NumWrites(), w2.NumWrites(), mqttClient.IsConnected(), mqttClient.IsConnectionOpen(), pgOk)
	})
	if err := http.ListenAndServe(":3002", router); err != nil {
		logger.Error("error listening", zap.Error(err))
	}
}
