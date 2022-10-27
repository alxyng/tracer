package simulation

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Solar struct {
	sync.Mutex

	MinSolarVoltage    float64
	MaxSolarVoltage    float64
	MinResistance      float64
	MaxResistance      float64
	MaxDeltaVoltage    float64
	MaxDeltaResistance float64
	TicksPerSecond     int

	solarVoltage float64
	solarCurrent float64
	resistance   float64
}

func (s *Solar) Run(ctx context.Context) {
	s.init()
	go s.loop(ctx)
}

func (s *Solar) init() {
	s.Lock()
	defer s.Unlock()

	if s.MinSolarVoltage < 0 {
		s.MinSolarVoltage = 0
	}

	if s.MinResistance < 0 {
		s.MinResistance = 0
	}

	s.solarVoltage = rand.Float64()*(s.MaxSolarVoltage-s.MinSolarVoltage) + s.MinSolarVoltage
	s.resistance = rand.Float64()*(s.MaxResistance-s.MinResistance) + s.MinResistance
	s.solarCurrent = s.solarVoltage / s.resistance
}

func (s *Solar) loop(ctx context.Context) {
	ticker := time.NewTicker((1000 / time.Duration(s.TicksPerSecond)) * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			s.tick(t)
		}
	}
}

func (s *Solar) tick(t time.Time) {
	s.Lock()
	defer s.Unlock()

	s.solarVoltage += (rand.Float64() * (2 * s.MaxDeltaVoltage)) - s.MaxDeltaVoltage
	if s.solarVoltage > s.MaxSolarVoltage {
		s.solarVoltage = s.MaxSolarVoltage
	} else if s.solarVoltage < s.MinSolarVoltage {
		s.solarVoltage = s.MinSolarVoltage
	}

	s.resistance += (rand.Float64() * (2 * s.MaxDeltaResistance)) - s.MaxDeltaResistance
	if s.resistance > s.MaxResistance {
		s.resistance = s.MaxResistance
	} else if s.resistance < s.MinResistance {
		s.resistance = s.MinResistance
	}

	s.solarCurrent = s.solarVoltage / s.resistance
}

type Measurement struct {
	Time    time.Time
	Voltage float64
	Current float64
}

func (s *Solar) GetMeasurement() Measurement {
	s.Lock()
	defer s.Unlock()

	return Measurement{time.Now().UTC(), s.solarVoltage, s.solarCurrent}
}
