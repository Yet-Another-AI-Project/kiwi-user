package infrastructure

import (
	"crypto/tls"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/infrastructure/jwt"
	"kiwi-user/internal/infrastructure/payment/stripe"
	"kiwi-user/internal/infrastructure/repository"
	"net/http"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/alibaba/captcha"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/cache"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/resend"
	"github.com/Yet-Another-AI-Project/kiwi-lib/client/volcengine/msgsms"

	"github.com/posthog/posthog-go"
	"go.uber.org/fx"

	"github.com/futurxlab/golanggraph/logger"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/alibaba/oss"
	liblogger "github.com/Yet-Another-AI-Project/kiwi-lib/logger"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/xhttp"
	"github.com/futurxlab/golanggraph/xerror"
)

func newLogger(config *config.Config) (logger.ILogger, error) {
	logger, err := liblogger.NewLogger(
		liblogger.WithLevel(config.Log.Level),
		liblogger.WithFilePath(config.Log.File),
		liblogger.WithFormat(config.Log.Format),
	)

	if err != nil {
		return nil, xerror.Wrap(err)
	}
	return logger, nil
}

func newPosthogClient(config *config.Config, logger logger.ILogger) (posthog.Client, error) {
	posthogClient, err := posthog.NewWithConfig(
		config.Posthog.ProjectAPIKey,
		posthog.Config{
			PersonalApiKey: config.Posthog.PersonalAPIKey,
			Endpoint:       config.Posthog.Endpoint,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Verbose: false,
		},
	)

	if err != nil {
		return nil, err
	}

	return posthogClient, err
}

func newHttpClient() *xhttp.Client {
	return xhttp.NewClient(xhttp.WithTimeout(5 * time.Second))
}

// new sms
func newSmsClient(config *config.Config) msgsms.SmsClient {
	smsClient := msgsms.NewSmsClient(
		msgsms.WithAccessKey(config.Sms.AccessKey),
		msgsms.WithSecretKey(config.Sms.SecretKey),
		msgsms.WithSmsAccount(config.Sms.SmsAccount),
		msgsms.WithSignName(config.Sms.SignName),
		msgsms.WithDefaultScene(config.Sms.DefaultScene),
		msgsms.WithTemplateID(config.Sms.VerifyTemplateID),
	)
	return smsClient
}

func newCaptchaClient(config *config.Config) captcha.CaptchaClient {
	captchaClient, err := captcha.NewCaptchaClient(
		captcha.WithAccessKeyId(config.Captcha.AccessKeyID),
		captcha.WithAccessKeySecret(config.Captcha.AccessKeySecret),
	)
	if err != nil {
		return nil
	}
	return captchaClient
}

func newMailClient(config *config.Config) *resend.ResendClient {
	client := resend.NewResendClient(
		resend.WithAPIKey(config.Mail.ResendAPIKey),
		resend.WithFrom(config.Mail.ResendFromEmail),
		resend.WithVerifyCodeTemplate(config.Mail.VertifyCodeTemplate),
		resend.WithVerifyCodeSubject(config.Mail.VertifyCodeSubject),
	)
	return client
}

func newOSSClient(cfg *config.Config) (*oss.AliyunOss, error) {
	if cfg.OSS == nil {
		return nil, xerror.New("OSS config is not provided")
	}
	return oss.NewAliyunOss(cfg.OSS.Endpoint, cfg.OSS.AccessKeyID, cfg.OSS.AccessKeySecret)
}

func newStripeClient(cfg *config.Config) *stripe.StripeClient {
	return stripe.NewStripeClient(
		cfg.Payment.StripeAPIKey,
		cfg.Payment.StripeWebhookSecret,
		cfg.Payment.StripeSuccessURL,
		cfg.Payment.StripeCancelURL,
		cfg.Payment.StripeMonthlyPriceID,
		cfg.Payment.StripeYearlyPriceID,
	)
}

var Module = fx.Provide(
	// oss client
	newOSSClient,

	// logger
	newLogger,

	// http client
	newHttpClient,

	// posthog client
	fx.Annotate(
		newPosthogClient,
		fx.OnStop(func(client posthog.Client) {
			client.Close()
		}),
	),

	// jwt
	jwt.NewRSA,
	jwt.NewJWTHelper,

	// initialize the repository modules
	fx.Annotate(
		repository.NewClient,
		fx.OnStop(func(client *repository.Client) {
			client.Close()
		}),
	),

	fx.Annotate(
		repository.NewUserImpl,
		fx.As(new(contract.IUserRepository)),
		fx.As(new(contract.IUserReadRepository)),
		fx.As(new(contract.IUserWriteRepository)),
	),

	fx.Annotate(
		repository.NewApplicationImpl,
		fx.As(new(contract.IApplicationRepository)),
		fx.As(new(contract.IApplicationReadRepository)),
		fx.As(new(contract.IApplicationWriteRepository)),
	),

	fx.Annotate(
		repository.NewDeviceImpl,
		fx.As(new(contract.IDeviceRepository)),
		fx.As(new(contract.IDeviceReadRepository)),
		fx.As(new(contract.IDeviceWriteRepository)),
	),

	fx.Annotate(
		repository.NewRoleImpl,
		fx.As(new(contract.IRoleRepository)),
		fx.As(new(contract.IRoleReadRepository)),
		fx.As(new(contract.IRoleWriteRepository)),
	),

	fx.Annotate(
		repository.NewOrganizationUserImpl,
		fx.As(new(contract.IOrganizationUserRepository)),
		fx.As(new(contract.IOrganizationUserReadRepository)),
		fx.As(new(contract.IOrganizationUserWriteRepository)),
	),

	fx.Annotate(
		repository.NewOrganizationImpl,
		fx.As(new(contract.IOrganizationRepository)),
		fx.As(new(contract.IOrganizationReadRepository)),
		fx.As(new(contract.IOrganizationWriteRepository)),
	),

	fx.Annotate(
		repository.NewOrganizationApplicationImpl,
		fx.As(new(contract.IOrganizationApplicationRepository)),
		fx.As(new(contract.IOrganizationApplicationReadRepository)),
		fx.As(new(contract.IOrganizationApplicationWriteRepository)),
	),

	fx.Annotate(
		repository.NewBindingVerifyImpl,
		fx.As(new(contract.IBindingVerifyRepository)),
		fx.As(new(contract.IBindingVerifyReadRepository)),
		fx.As(new(contract.IBindingVerifyWriteRepository)),
	),

	fx.Annotate(
		repository.NewPaymentImpl,
		fx.As(new(contract.IPaymentRepository)),
		fx.As(new(contract.IPaymentReadRepository)),
		fx.As(new(contract.IPaymentWriteRepository)),
	),

	fx.Annotate(
		repository.NewStripeEventImpl,
		fx.As(new(contract.IStripeEventRepository)),
		fx.As(new(contract.IStripeEventReadRepository)),
		fx.As(new(contract.IStripeEventWriteRepository)),
	),

	fx.Annotate(
		repository.NewMailVertifyCodeImpl,
		fx.As(new(contract.IMailVertifyCodeRepository)),
		fx.As(new(contract.IMailVertifyCodeReadRepository)),
		fx.As(new(contract.IMailVertifyCodeWriteRepository)),
	),

	// sms
	newSmsClient,

	// mail
	newMailClient,

	// captcha
	newCaptchaClient,
	cache.NewMemCache,

	// stripe client
	newStripeClient,
)
