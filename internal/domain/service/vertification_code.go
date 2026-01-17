package service

import (
	"context"
	"errors"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/utils"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/resend"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

type VertificationCodeService struct {
	mailVertifyCodeRepository contract.IMailVertifyCodeRepository
	mailClient                *resend.ResendClient
	logger                    logger.ILogger
}

func NewVertificationCodeService(
	logger logger.ILogger,
	mailClient *resend.ResendClient,
	mailVertifyCodeRepository contract.IMailVertifyCodeRepository) *VertificationCodeService {

	return &VertificationCodeService{
		logger:                    logger,
		mailClient:                mailClient,
		mailVertifyCodeRepository: mailVertifyCodeRepository,
	}
}

// SendEmailVerificationCode sends a verification code to the specified email
func (v *VertificationCodeService) SendEmailVerificationCode(
	ctx context.Context,
	email string,
	codetype enum.VertificationCodeType,
) error {
	// 1. Generate a random verification code
	code := utils.GenerateVerificationCode()

	// 2. Check if verification code exists for this email
	existingCode, err := v.mailVertifyCodeRepository.Find(ctx, email, codetype)
	if err != nil {
		return xerror.Wrap(err)
	}

	// 3. Create or update verification code
	mailVertifyCode := &entity.MailVertifyCodeEntity{
		Email:     email,
		Code:      code,
		Type:      codetype,
		ExpiresAt: time.Now().Add(time.Minute * 10),
	}

	if existingCode != nil {
		// Update existing code
		mailVertifyCode.ID = existingCode.ID // 设置ID
		err = v.mailVertifyCodeRepository.Update(ctx, mailVertifyCode)
	} else {
		// Create new code
		err = v.mailVertifyCodeRepository.Create(ctx, mailVertifyCode)
	}

	if err != nil {
		return xerror.Wrap(err)
	}

	// 4. Send the code via email
	err = v.mailClient.EnqueueVerifyCode(email, code)
	if err != nil {
		v.logger.Errorf(ctx, "Failed to enqueue verification code: %w", err)
		return xerror.Wrap(err)
	}

	return nil
}

func (v *VertificationCodeService) VerifyEmailCode(
	ctx context.Context,
	email string,
	code string,
	codetype enum.VertificationCodeType,
) (bool, error) {
	existingCode, err := v.mailVertifyCodeRepository.Find(ctx, email, codetype)
	if err != nil {
		return false, xerror.Wrap(err)
	}

	if existingCode == nil {
		return false, xerror.Wrap(errors.New("verification code not found"))
	}

	if existingCode.ExpiresAt.Before(time.Now()) {
		return false, xerror.Wrap(errors.New("verification code expired"))
	}

	if existingCode.Code != code {
		return false, xerror.Wrap(errors.New("invalid verification code"))
	}

	err = v.mailVertifyCodeRepository.Delete(ctx, existingCode)
	if err != nil {
		return false, xerror.Wrap(err)
	}

	return true, nil
}
