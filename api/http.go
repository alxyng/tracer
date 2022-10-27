package api

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alxyng/tracer/controller"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func NewHTTPTransport(router *mux.Router, mqttClient mqtt.Client, logger *zap.Logger) *HTTPTransport {
	return &HTTPTransport{
		router:     router,
		mqttClient: mqttClient,
		logger:     logger,
	}
}

type HTTPTransport struct {
	router     *mux.Router
	mqttClient mqtt.Client
	logger     *zap.Logger
}

func (t *HTTPTransport) Register() {
	t.router.HandleFunc("/reading", t.createHandler(controller.TopicGetReading)).Methods(http.MethodPost)
	t.router.HandleFunc("/getTime", t.createHandler(controller.TopicGetSystemTime)).Methods(http.MethodPost)
	t.router.HandleFunc("/setTime", t.createHandler(controller.TopicSetSystemTime)).Methods(http.MethodPost)
	t.router.HandleFunc("/getBatteryInformation", t.createHandler(controller.TopicGetBatteryInformation)).Methods(http.MethodPost)
	t.router.HandleFunc("/setBatteryCapacity", t.createHandler(controller.TopicSetBatteryCapacity)).Methods(http.MethodPost)
}

func (t *HTTPTransport) createHandler(topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		done := make(chan []byte)

		responseTopic := strings.Replace(topic, "request/#", "response/"+requestID, -1)
		if token := t.mqttClient.Subscribe(responseTopic, 0, func(c mqtt.Client, msg mqtt.Message) {
			done <- msg.Payload()
		}); token.Wait() && token.Error() != nil {
			t.logger.Error("error subscribing to response topic", zap.String("topic", responseTopic), zap.Error(token.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer func() {
			if token := t.mqttClient.Unsubscribe(responseTopic); token.Wait() && token.Error() != nil {
				t.logger.Error("error unsubscribing to response topic", zap.String("topic", responseTopic), zap.Error(token.Error()))
			}
		}()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.logger.Error("error reading body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		requestTopic := strings.Replace(topic, "#", requestID, -1)
		if token := t.mqttClient.Publish(requestTopic, 0, false, data); token.Wait() && token.Error() != nil {
			t.logger.Error("error publishing to mqtt", zap.String("topic", "tracer/reading"), zap.Error(token.Error()))
			return
		}

		select {
		case data := <-done:
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		case <-time.After(5 * time.Second):
			t.logger.Error("timed out waiting for mqtt reply")
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		}
	}
}
