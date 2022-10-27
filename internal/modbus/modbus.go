package modbus

import (
	"time"

	"github.com/goburrow/modbus"
)

func Connect(cfg *Config) (*modbus.RTUClientHandler, modbus.Client, error) {
	handler := modbus.NewRTUClientHandler(cfg.Addr)
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = cfg.SlaveID
	handler.Timeout = 5 * time.Second

	if err := handler.Connect(); err != nil {
		return nil, nil, err
	}

	return handler, modbus.NewClient(handler), nil
}
