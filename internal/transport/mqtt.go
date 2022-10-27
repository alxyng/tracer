package transport

import (
	"context"
	"encoding/json"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type ServiceHandler[Req, Res any] func(context.Context, *Req) (*Res, error)

func MQTT[Req, Res any](c mqtt.Client, logger *zap.Logger, handler ServiceHandler[Req, Res]) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		var err error
		var req Req
		var res *Res

		if err := json.Unmarshal(msg.Payload(), &req); err != nil {
			logger.Error("error unmarshalling request", zap.Error(err))
			publishError(c, logger, msg.Topic(), "error unmarshalling request")
			return
		}

		res, err = handler(context.Background(), &req)
		if err != nil {
			logger.Error("error handling request", zap.Error(err))
			publishError(c, logger, msg.Topic(), "eerror handling request")
			return
		}

		payload, err := json.Marshal(&res)
		if err != nil {
			logger.Error("error marshalling response", zap.Error(err))
			publishError(c, logger, msg.Topic(), "error marshalling response")
			return
		}

		publishData(c, logger, msg.Topic(), payload)
	}
}

func publishData(c mqtt.Client, logger *zap.Logger, topic string, data []byte) error {
	topic = strings.Replace(topic, "request", "response", 1)
	if token := c.Publish(topic, 0, false, data); token.Wait() && token.Error() != nil {
		logger.Error("error publishing data", zap.String("topic", topic), zap.Error(token.Error()))
		return token.Error()
	}
	return nil
}

func publishError(c mqtt.Client, logger *zap.Logger, topic string, err string) error {
	res := `{"error": "` + err + `"}`
	topic = strings.Replace(topic, "request", "response", 1)
	if token := c.Publish(topic, 0, false, res); token.Wait() && token.Error() != nil {
		logger.Error("error publishing error", zap.String("topic", topic), zap.Error(token.Error()))
		return token.Error()
	}
	return nil
}
