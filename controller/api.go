package controller

import (
	"context"
)

type API interface {
	Run(ctx context.Context)
	OnRecord(f func(context.Context, *Reading))

	GetReading(ctx context.Context, req *GetReadingRequest) (*GetReadingResponse, error)
	GetSystemTime(ctx context.Context, req *GetSystemTimeRequest) (*GetSystemTimeResponse, error)
	SetSystemTime(ctx context.Context, req *SetSystemTimeRequest) (*SetSystemTimeResponse, error)
	GetBatteryInformation(ctx context.Context, req *GetBatteryInformationRequest) (*GetBatteryInformationResponse, error)
	SetBatteryCapacity(ctx context.Context, req *SetBatteryCapacityRequest) (*SetBatteryCapacityResponse, error)
}
