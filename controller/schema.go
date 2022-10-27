package controller

import (
	"encoding/json"
	"time"
)

type Reading struct {
	OverTemperature bool `json:"overTemperature"` // A1
	Day             bool `json:"day"`             // A2

	SolarVoltage float32 `json:"solarVoltage"` // A3
	SolarCurrent float32 `json:"solarCurrent"` // A4
	SolarPower   float32 `json:"solarPower"`   // A5, A6

	LoadVoltage float32 `json:"loadVoltage"` // A7
	LoadCurrent float32 `json:"loadCurrent"` // A8
	LoadPower   float32 `json:"loadPower"`   // A9, A10

	BatteryTemperature  float32 `json:"batteryTemperature"`  // A11
	DeviceTemperature   float32 `json:"deviceTemperature"`   // A12
	BatterySOC          uint16  `json:"batterySOC"`          // A13
	BatteryRatedVoltage uint16  `json:"batteryRatedVoltage"` // A14

	// BatteryStatus              any // A15
	// ChargingEquipmentStatus    any // A15
	// DischargingEquipmentStatus any // A15

	MaximumBatteryVoltageToday float32 `json:"maximumBatteryVoltageToday"` // A18
	MinimumBatteryVoltageToday float32 `json:"minimumBatteryVoltageToday"` // A19

	ConsumedEnergyToday float32 `json:"consumedEnergyToday"` // A20, A21
	ConsumedEnergyMonth float32 `json:"consumedEnergyMonth"` // A22, A23
	ConsumedEnergyYear  float32 `json:"consumedEnergyYear"`  // A24, A25
	ConsumedEnergyTotal float32 `json:"consumedEnergyTotal"` // A26, A27

	GeneratedEnergyToday float32 `json:"generatedEnergyToday"` // A28, A29
	GeneratedEnergyMonth float32 `json:"generatedEnergyMonth"` // A30, A31
	GeneratedEnergyYear  float32 `json:"generatedEnergyYear"`  // A32, A33
	GeneratedEnergyTotal float32 `json:"generatedEnergyTotal"` // A34, A35

	BatteryVoltage float32 `json:"batteryVoltage"` // A36
	BatteryCurrent float32 `json:"batteryCurrent"` // A37, A38

	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"endTime"`
	Duration  time.Duration `json:"duration"`
}

type BatteryType int

const (
	BatteryTypeUserDefined BatteryType = iota
	BatteryTypeSealed
	BatteryTypeGel
	BatteryTypeFlooded
)

func (bt BatteryType) String() string {
	switch bt {
	case BatteryTypeUserDefined:
		return "User defined"
	case BatteryTypeSealed:
		return "Sealed"
	case BatteryTypeGel:
		return "Gel"
	case BatteryTypeFlooded:
		return "Flooded"
	}
	return "Unknown"
}

func (bt BatteryType) MarshalJSON() ([]byte, error) {
	return json.Marshal(bt.String())
}

type GetReadingRequest struct {
}

type GetReadingResponse struct {
	Reading Reading `json:"reading"`
}

type GetSystemTimeRequest struct{}

type GetSystemTimeResponse struct {
	Time time.Time `json:"time"`
}

type SetSystemTimeRequest struct {
	Time time.Time `json:"time"`
}

type SetSystemTimeResponse struct{}

type GetBatteryInformationRequest struct{}

type GetBatteryInformationResponse struct {
	BatteryType     BatteryType `json:"batteryType"`
	BatteryCapacity uint16      `json:"batteryCapacity"`
}

type SetBatteryCapacityRequest struct {
	Capacity uint16 `json:"capacity"`
}

type SetBatteryCapacityResponse struct{}
