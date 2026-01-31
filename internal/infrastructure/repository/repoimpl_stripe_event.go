package repository

import (
	"context"
	"time"

	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/stripeevent"
)

type stripeEventImpl struct {
	baseImpl
}

func (s *stripeEventImpl) FindByEventID(ctx context.Context, eventID string) (*entity.StripeEventEntity, error) {
	db := s.getEntClient(ctx)

	eventDO, err := db.StripeEvent.Query().Where(stripeevent.EventID(eventID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return convertStripeEventDOToEntity(eventDO), nil
}

func (s *stripeEventImpl) ExistsByEventID(ctx context.Context, eventID string) (bool, error) {
	db := s.getEntClient(ctx)

	exists, err := db.StripeEvent.Query().Where(stripeevent.EventID(eventID)).Exist(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *stripeEventImpl) Create(ctx context.Context, event *entity.StripeEventEntity) (*entity.StripeEventEntity, error) {
	db := s.getEntClient(ctx)

	eventDO, err := db.StripeEvent.Create().
		SetEventID(event.EventID).
		SetEventType(event.EventType).
		SetProcessed(event.Processed).
		SetCreatedAt(event.CreatedAt).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return convertStripeEventDOToEntity(eventDO), nil
}

func (s *stripeEventImpl) MarkProcessed(ctx context.Context, eventID string) error {
	db := s.getEntClient(ctx)

	_, err := db.StripeEvent.Update().
		Where(stripeevent.EventID(eventID)).
		SetProcessed(true).
		SetProcessedAt(time.Now()).
		Save(ctx)

	return err
}

func NewStripeEventImpl(db *Client) contract.IStripeEventRepository {
	return &stripeEventImpl{
		baseImpl{
			db: db,
		},
	}
}

func convertStripeEventDOToEntity(event *ent.StripeEvent) *entity.StripeEventEntity {
	return &entity.StripeEventEntity{
		ID:          event.ID,
		EventID:     event.EventID,
		EventType:   event.EventType,
		Processed:   event.Processed,
		ProcessedAt: event.ProcessedAt,
		CreatedAt:   event.CreatedAt,
	}
}
