package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
)

type IDeviceReadRepository interface {
	FindByDevice(ctx context.Context, userID string, deviceType, deviceID string) (*aggregate.DeviceAggregate, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*aggregate.DeviceAggregate, error)
}

type IDeviceWriteRepository interface {
	Create(ctx context.Context, device *aggregate.DeviceAggregate) (*aggregate.DeviceAggregate, error)
	Update(ctx context.Context, device *aggregate.DeviceAggregate) (*aggregate.DeviceAggregate, error)
}

type IDeviceRepository interface {
	ITransaction
	IDeviceReadRepository
	IDeviceWriteRepository
}
