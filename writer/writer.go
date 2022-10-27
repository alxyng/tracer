package writer

import (
	"context"

	"github.com/alxyng/tracer/controller"
	"github.com/jackc/pgx/v5"
)

type Writer interface {
	Write(ctx context.Context, reading *controller.Reading) error
}

func NewSQLWriter(conn *pgx.Conn) *SQLWriter {
	return &SQLWriter{
		conn: conn,
	}
}

type SQLWriter struct {
	conn *pgx.Conn
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
	return err
}
