package config

import (
	"os"
	"strconv"

	"github.com/alxyng/tracer/internal/modbus"
)

const defaultAPIPort = "3001"
const defaultDatabaseDSN = "postgres://username:password@host:5432/database"
const defaultModbusAddr = "/dev/serial0"
const defaultModbusSlaveID = 1

type Config struct {
	API      *APIConfig
	Database *DatabaseConfig
	Modbus   *modbus.Config
}

type APIConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

func Get() (*Config, error) {
	cfg := &Config{
		API: &APIConfig{
			Port: defaultAPIPort,
		},
		Database: &DatabaseConfig{
			DSN: defaultDatabaseDSN,
		},
		Modbus: &modbus.Config{
			Addr:    defaultModbusAddr,
			SlaveID: defaultModbusSlaveID,
		},
	}

	if apiPort := os.Getenv("TRACER_API_PORT"); apiPort != "" {
		cfg.API.Port = apiPort
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

	return cfg, nil
}
