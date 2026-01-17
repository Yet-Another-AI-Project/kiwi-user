package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/device"

	"github.com/futurxlab/golanggraph/xerror"
)

type deviceImpl struct {
	baseImpl
}

func (d *deviceImpl) FindByDevice(ctx context.Context, userID string, deviceType, deviceID string) (*aggregate.DeviceAggregate, error) {
	db := d.getEntClient(ctx)

	deviceDO, err := db.Device.Query().Where(device.UserID(userID), device.DeviceType(deviceType), device.DeviceID(deviceID)).Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if deviceDO == nil {
		return nil, nil
	}

	userDO, err := deviceDO.QueryUser().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.DeviceAggregate{
		Device: convertDeviceDOToEntity(deviceDO),
		User:   convertUserDOToEntity(userDO),
	}, nil
}

func (d *deviceImpl) FindByRefreshToken(ctx context.Context, refreshToken string) (*aggregate.DeviceAggregate, error) {
	panic("not implemented")
}

func (d *deviceImpl) Create(ctx context.Context, device *aggregate.DeviceAggregate) (*aggregate.DeviceAggregate, error) {
	db := d.getEntClient(ctx)

	deviceDO, err := db.Device.Create().
		SetDeviceType(device.Device.DeviceType).
		SetDeviceID(device.Device.DeviceID).
		SetRefreshToken(device.Device.RefreshToken).
		SetRefreshTokenExpiresAt(device.Device.RefreshTokenExpiresAt).
		SetUserID(device.User.ID).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	device.Device = convertDeviceDOToEntity(deviceDO)

	return device, nil
}

func (d *deviceImpl) Update(ctx context.Context, device *aggregate.DeviceAggregate) (*aggregate.DeviceAggregate, error) {
	db := d.getEntClient(ctx)

	deviceDO, err := db.Device.UpdateOneID(device.Device.ID).
		SetRefreshToken(device.Device.RefreshToken).
		SetRefreshTokenExpiresAt(device.Device.RefreshTokenExpiresAt).
		SetUserID(device.User.ID).
		SetOrganizationID(device.Device.OrganizationID).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	device.Device = convertDeviceDOToEntity(deviceDO)

	return device, nil
}

func NewDeviceImpl(db *Client) contract.IDeviceRepository {
	return &deviceImpl{
		baseImpl{
			db: db,
		},
	}
}
