package config

import (
	"os"
	"strconv"

	"github.com/alxyng/tracer/internal/modbus"
)

const defaultAPIAddr = ":3001"
const defaultDatabaseDSN = "postgres://username:password@host:5432/database"
const defaultModbusAddr = "/dev/serial0"
const defaultModbusSlaveID = 1
const defaultMQTTBroker = "tcp://localhost:1883"

type Config struct {
	API      *APIConfig
	Database *DatabaseConfig
	Modbus   *modbus.Config
	MQTT     *MQTTConfig
}

type APIConfig struct {
	Addr string
}

type DatabaseConfig struct {
	DSN string
}

type MQTTConfig struct {
	Broker string
}

func Get() (*Config, error) {
	cfg := &Config{
		API: &APIConfig{
			Addr: defaultAPIAddr,
		},
		Database: &DatabaseConfig{
			DSN: defaultDatabaseDSN,
		},
		Modbus: &modbus.Config{
			Addr:    defaultModbusAddr,
			SlaveID: defaultModbusSlaveID,
		},
		MQTT: &MQTTConfig{
			Broker: defaultMQTTBroker,
		},
	}

	if apiAddr := os.Getenv("TRACER_API_ADDR"); apiAddr != "" {
		cfg.API.Addr = apiAddr
	}

	if databaseDSN := os.Getenv("TRACER_DATABASE_DSN"); databaseDSN != "" {
		cfg.Database.DSN = databaseDSN
	}

	if modbusAddr := os.Getenv("TRACER_MODBUS_ADDR"); modbusAddr != "" {
		cfg.Modbus.Addr = modbusAddr
	}

	if modbusSlaveID := os.Getenv("TRACER_MODBUS_SLAVEID"); modbusSlaveID != "" {
		slaveID, err := strconv.ParseUint(modbusSlaveID, 10, 8)
		if err != nil {
			return nil, err
		}
		cfg.Modbus.SlaveID = byte(slaveID)
	}

	if mqttBroker := os.Getenv("TRACER_MQTT_BROKER"); mqttBroker != "" {
		cfg.MQTT.Broker = mqttBroker
	}

	return cfg, nil
}
