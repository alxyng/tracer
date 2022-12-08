package writer

import (
	"context"
	"sync/atomic"

	"github.com/alxyng/tracer/controller"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Writer interface {
	Write(ctx context.Context, reading *controller.Reading) error
	NumWrites() uint64
}

func NewSQLWriter(conn *pgx.Conn, logger *zap.Logger) *SQLWriter {
	return &SQLWriter{
		conn:   conn,
		logger: logger,
		writes: 0,
	}
}

type SQLWriter struct {
	conn   *pgx.Conn
	logger *zap.Logger
	writes uint64
}

func (w *SQLWriter) Write(ctx context.Context, reading controller.Reading) error {
	_, err := w.conn.Exec(ctx, `
			INSERT INTO readings VALUES(
				default, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
				$21, $22, $23, $24, $25, $26);`,
		reading.OverTemperature,
		reading.Day,
		reading.SolarVoltage,
		reading.SolarCurrent,
		reading.SolarPower,
		reading.LoadVoltage,
		reading.LoadCurrent,
		reading.LoadPower,
		reading.BatteryTemperature,
		reading.DeviceTemperature,
		reading.BatterySOC,
		reading.BatteryRatedVoltage,
		reading.MaximumBatteryVoltageToday,
		reading.MinimumBatteryVoltageToday,
		reading.ConsumedEnergyToday,
		reading.ConsumedEnergyMonth,
		reading.ConsumedEnergyYear,
		reading.ConsumedEnergyTotal,
		reading.GeneratedEnergyToday,
		reading.GeneratedEnergyMonth,
		reading.GeneratedEnergyYear,
		reading.GeneratedEnergyTotal,
		reading.BatteryVoltage,
		reading.BatteryCurrent,
		reading.Duration,
		reading.EndTime,
	)
	if err != nil {
		return err
	}

	atomic.AddUint64(&w.writes, 1)
	return nil
}

func (w *SQLWriter) NumWrites() uint64 {
	return atomic.LoadUint64(&w.writes)
}

func NewSQLAggregateWriter(conn *pgx.Conn, logger *zap.Logger) *SQLAggregateWriter {
	return &SQLAggregateWriter{
		conn:   conn,
		logger: logger,
		writes: 0,
	}
}

type SQLAggregateWriter struct {
	conn   *pgx.Conn
	logger *zap.Logger
	last   controller.Reading
	writes uint64
}

func (w *SQLAggregateWriter) Write(ctx context.Context, reading controller.Reading) error {
	defer func() {
		w.last = reading
	}()

	if !w.hasLastReading() {
		return nil
	}

	dt := reading.EndTime.Sub(w.last.EndTime)

	avgPowerSolar := (reading.SolarPower + w.last.SolarPower) / 2
	generatedEnergy := dt.Seconds() * float64(avgPowerSolar)

	avgPowerLoad := (reading.LoadPower + w.last.LoadPower) / 2
	consumedEnergy := dt.Seconds() * float64(avgPowerLoad)

	_, err := w.conn.Exec(ctx, `
			INSERT INTO daily_energy (generated_energy, consumed_energy, time)
			VALUES($1, $2, date_trunc('day', $3::timestamp))
			ON CONFLICT (time) DO UPDATE
			SET generated_energy = daily_energy.generated_energy + EXCLUDED.generated_energy,
				consumed_energy = daily_energy.consumed_energy + EXCLUDED.consumed_energy;`,
		generatedEnergy,
		consumedEnergy,
		reading.EndTime,
	)
	if err != nil {
		return err
	}

	atomic.AddUint64(&w.writes, 1)
	return nil
}

func (w *SQLAggregateWriter) NumWrites() uint64 {
	return atomic.LoadUint64(&w.writes)
}

func (w *SQLAggregateWriter) hasLastReading() bool {
	return !w.last.EndTime.IsZero()
}
