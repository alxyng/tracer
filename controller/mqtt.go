package controller

import (
	"github.com/alxyng/tracer/internal/transport"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

const TopicGetReading = "tracer/controller/getReading/request/#"
const TopicGetSystemTime = "tracer/controller/getSystemTime/request/#"
const TopicSetSystemTime = "tracer/controller/setSystemTime/request/#"
const TopicGetBatteryInformation = "tracer/controller/getBatteryInformation/request/#"
const TopicSetBatteryCapacity = "tracer/controller/setBatteryCapacity/request/#"

func NewMQTTTransport(mqttClient mqtt.Client, api API, logger *zap.Logger) *MQTTTransport {
	return &MQTTTransport{
		mqttClient: mqttClient,
		api:        api,
		logger:     logger,
	}
}

type MQTTTransport struct {
	mqttClient mqtt.Client
	api        API
	logger     *zap.Logger
}

func (t *MQTTTransport) Register() error {
	if err := t.subscribe(TopicGetReading, 0,
		transport.MQTT(t.mqttClient, t.logger, t.api.GetReading)); err != nil {
		return err
	}

	if err := t.subscribe(TopicGetSystemTime, 0,
		transport.MQTT(t.mqttClient, t.logger, t.api.GetSystemTime)); err != nil {
		return err
	}

	if err := t.subscribe(TopicSetSystemTime, 0,
		transport.MQTT(t.mqttClient, t.logger, t.api.SetSystemTime)); err != nil {
		return err
	}

	if err := t.subscribe(TopicGetBatteryInformation, 0,
		transport.MQTT(t.mqttClient, t.logger, t.api.GetBatteryInformation)); err != nil {
		return err
	}

	if err := t.subscribe(TopicSetBatteryCapacity, 0,
		transport.MQTT(t.mqttClient, t.logger, t.api.SetBatteryCapacity)); err != nil {
		return err
	}

	return nil
}

func (t *MQTTTransport) subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	if token := t.mqttClient.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
