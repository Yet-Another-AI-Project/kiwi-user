package service

import (
	"context"
	"fmt"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/infrastructure/utils"
	"time"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

type DeviceService struct {
	deviceRepository         contract.IDeviceRepository
	refreshTokenExpireSecond int64
}

func NewDeviceService(config *config.Config, deviceRepository contract.IDeviceRepository) *DeviceService {
	return &DeviceService{
		deviceRepository:         deviceRepository,
		refreshTokenExpireSecond: config.JWT.RefreshTokenExpireSecond,
	}
}

func (d *DeviceService) UpsertDevice(
	ctx context.Context,
	userID string,
	deviceType string,
	deviceID string,
	organizationID uuid.UUID) (*aggregate.DeviceAggregate, error) {

	deviceAggregate, err := d.deviceRepository.FindByDevice(ctx, userID, deviceType, deviceID)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if deviceAggregate == nil {

		deviceAggregate = &aggregate.DeviceAggregate{
			Device: &entity.DeviceEntity{
				DeviceType: deviceType,
				DeviceID:   deviceID,
				RefreshToken: utils.GenerateRefreshToken(fmt.Sprintf("%s%s%s",
					userID,
					deviceType,
					deviceID)),
				RefreshTokenExpiresAt: time.Now().Add(time.Duration(d.refreshTokenExpireSecond) * time.Second),
				OrganizationID:        organizationID,
			},
			User: &entity.UserEntity{
				ID: userID,
			},
		}

		deviceAggregate, err = d.deviceRepository.Create(ctx, deviceAggregate)

		if err != nil {
			return nil, xerror.Wrap(err)
		}
	} else {
		deviceAggregate.Device.RefreshToken = utils.GenerateRefreshToken(fmt.Sprintf("%s%s%s",
			userID,
			deviceType,
			deviceID))
		deviceAggregate.Device.RefreshTokenExpiresAt = time.Now().Add(time.Duration(d.refreshTokenExpireSecond) * time.Second)
		deviceAggregate.Device.OrganizationID = organizationID

		deviceAggregate, err = d.deviceRepository.Update(ctx, deviceAggregate)
		if err != nil {
			return nil, xerror.Wrap(err)
		}
	}

	return deviceAggregate, nil
}

func (d *DeviceService) RegenerateRefreshToken(ctx context.Context, deviceAggregate *aggregate.DeviceAggregate, refreshExpireTime bool) (*aggregate.DeviceAggregate, error) {

	deviceAggregate.Device.RefreshToken = utils.GenerateRefreshToken(fmt.Sprintf("%s%s%s",
		deviceAggregate.User.ID,
		deviceAggregate.Device.DeviceType,
		deviceAggregate.Device.DeviceID))

	if refreshExpireTime {
		deviceAggregate.Device.RefreshTokenExpiresAt = time.Now().Add(time.Duration(d.refreshTokenExpireSecond) * time.Second)
	}

	var err error
	deviceAggregate, err = d.deviceRepository.Update(ctx, deviceAggregate)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return deviceAggregate, nil
}

func (d *DeviceService) UpdateDevice(ctx context.Context, deviceAggregate *aggregate.DeviceAggregate) (*aggregate.DeviceAggregate, error) {

	deviceAggregate, err := d.deviceRepository.Update(ctx, deviceAggregate)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return deviceAggregate, nil
}
