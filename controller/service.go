package controller

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/goburrow/modbus"
	"go.uber.org/zap"
)

var (
	ErrNotEnoughData = errors.New("not enough data")
)

func NewService(client modbus.Client, logger *zap.Logger) *Service {
	return &Service{
		client:    client,
		logger:    logger,
		iteration: 0,
	}
}

type Service struct {
	readingMutex sync.Mutex
	serialMutex  sync.Mutex

	client    modbus.Client
	logger    *zap.Logger
	iteration uint64

	onRecord func(context.Context, *Reading)

	reading Reading
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case start := <-ticker.C:
			s.takeReading(ctx, start)
		}
	}
}

func (s *Service) OnRecord(f func(context.Context, *Reading)) {
	s.onRecord = f
}

func (s *Service) takeReading(ctx context.Context, start time.Time) {
	var results []byte
	var err error
	var reading Reading

	s.serialMutex.Lock()
	defer s.serialMutex.Unlock()

	reading.StartTime = start.UTC()

	results, err = s.client.ReadDiscreteInputs(0x2000, 1)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}
	reading.OverTemperature = results[0] != 0x00

	results, err = s.client.ReadDiscreteInputs(0x200c, 1)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}
	reading.Day = results[0] == 0x00

	results, err = s.client.ReadInputRegisters(0x3100, 4)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.SolarVoltage = getFloatFrom16Bit(results[0:2])
	reading.SolarCurrent = getFloatFrom16Bit(results[2:4])
	reading.SolarPower = getFloatFrom32Bit(results[4:8])

	results, err = s.client.ReadInputRegisters(0x310c, 4)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.LoadVoltage = getFloatFrom16Bit(results[0:2])
	reading.LoadCurrent = getFloatFrom16Bit(results[2:4])
	reading.LoadPower = getFloatFrom32Bit(results[4:8])

	results, err = s.client.ReadInputRegisters(0x3110, 2)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.BatteryTemperature = getFloatFrom16Bit(results[0:2])
	reading.DeviceTemperature = getFloatFrom16Bit(results[2:4])

	results, err = s.client.ReadInputRegisters(0x311a, 1)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.BatterySOC = getUint16(results[0:2])

	results, err = s.client.ReadInputRegisters(0x311d, 1)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.BatteryRatedVoltage = getUint16(results[0:2]) / 100

	results, err = s.client.ReadInputRegisters(0x3302, 18)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.MaximumBatteryVoltageToday = getFloatFrom16Bit(results[0:2])
	reading.MinimumBatteryVoltageToday = getFloatFrom16Bit(results[2:4])
	reading.ConsumedEnergyToday = getFloatFrom32Bit(results[4:8])
	reading.ConsumedEnergyMonth = getFloatFrom32Bit(results[8:12])
	reading.ConsumedEnergyYear = getFloatFrom32Bit(results[12:16])
	reading.ConsumedEnergyTotal = getFloatFrom32Bit(results[16:20])
	reading.GeneratedEnergyToday = getFloatFrom32Bit(results[20:24])
	reading.GeneratedEnergyMonth = getFloatFrom32Bit(results[24:28])
	reading.GeneratedEnergyYear = getFloatFrom32Bit(results[28:32])
	reading.GeneratedEnergyTotal = getFloatFrom32Bit(results[32:36])

	results, err = s.client.ReadInputRegisters(0x331a, 3)
	if err != nil {
		s.logger.Error("error reading", zap.Error(err))
		return
	}

	reading.BatteryVoltage = getFloatFrom16Bit(results[0:2])
	reading.BatteryCurrent = getFloatFrom32Bit(results[2:6])
	if reading.BatteryCurrent > 1000 {
		reading.BatteryCurrent = 0
	}

	reading.EndTime = time.Now().UTC()
	reading.Duration = reading.EndTime.Sub(reading.StartTime)

	s.readingMutex.Lock()
	defer s.readingMutex.Unlock()

	s.reading = reading

	if s.onRecord != nil {
		s.onRecord(ctx, &s.reading)
	}

	s.iteration++
}

func (s *Service) GetReading(ctx context.Context, req *GetReadingRequest) (*GetReadingResponse, error) {
	s.readingMutex.Lock()
	defer s.readingMutex.Unlock()

	return &GetReadingResponse{Reading: s.reading}, nil
}

func (s *Service) GetSystemTime(ctx context.Context, req *GetSystemTimeRequest) (*GetSystemTimeResponse, error) {
	s.serialMutex.Lock()
	defer s.serialMutex.Unlock()

	results, err := s.client.ReadHoldingRegisters(0x9013, 3)
	if err != nil {
		return nil, err
	}
	if len(results) < 6 {
		return nil, ErrNotEnoughData
	}

	min := int(results[0])
	sec := int(results[1])
	day := int(results[2])
	hour := int(results[3])
	year := 2000 + int(results[4])
	month := time.Month(results[5])

	return &GetSystemTimeResponse{Time: time.Date(year, month, day, hour, min, sec, 0, time.UTC)}, nil
}

func (s *Service) SetSystemTime(ctx context.Context, req *SetSystemTimeRequest) (*SetSystemTimeResponse, error) {
	s.serialMutex.Lock()
	defer s.serialMutex.Unlock()

	data := make([]byte, 6)

	data[0] = byte(req.Time.Minute())
	data[1] = byte(req.Time.Second())
	data[2] = byte(req.Time.Day())
	data[3] = byte(req.Time.Hour())
	data[4] = byte(req.Time.Year() - 2000)
	data[5] = byte(req.Time.Month())

	_, err := s.client.WriteMultipleRegisters(0x9013, 3, data)
	if err != nil {
		s.logger.Info("error setting system time", zap.Error(err))
		return nil, err
	}

	s.logger.Info("system time set", zap.Time("systemTime", req.Time))

	return &SetSystemTimeResponse{}, nil
}

func (s *Service) GetBatteryInformation(ctx context.Context, req *GetBatteryInformationRequest) (*GetBatteryInformationResponse, error) {
	s.serialMutex.Lock()
	defer s.serialMutex.Unlock()

	results, err := s.client.ReadHoldingRegisters(0x9000, 2)
	if err != nil {
		return nil, err
	}
	if len(results) < 4 {
		return nil, ErrNotEnoughData
	}

	return &GetBatteryInformationResponse{
		BatteryType:     BatteryType(results[1]),
		BatteryCapacity: getUint16(results[2:4]),
	}, nil
}

func (s *Service) SetBatteryCapacity(ctx context.Context, req *SetBatteryCapacityRequest) (*SetBatteryCapacityResponse, error) {
	s.serialMutex.Lock()
	defer s.serialMutex.Unlock()

	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, req.Capacity)

	_, err := s.client.WriteMultipleRegisters(0x9001, 1, data)
	if err != nil {
		s.logger.Info("error setting battery capacity", zap.Error(err))
		return nil, err
	}

	s.logger.Info("battery capacity set", zap.Uint16("capacity", req.Capacity))

	return &SetBatteryCapacityResponse{}, nil
}

func getUint16(data []byte) uint16 {
	return binary.BigEndian.Uint16(data)
}

func getFloatFrom16Bit(data []byte) float32 {
	return float32(binary.BigEndian.Uint16(data)) / 100
}

func getFloatFrom32Bit(data []byte) float32 {
	return float32(get32BitData(data)) / 100
}

func get32BitData(data []byte) uint32 {
	var buf []byte
	buf = append(buf, data[2:4]...)
	buf = append(buf, data[0:2]...)
	return binary.BigEndian.Uint32(buf)
}
