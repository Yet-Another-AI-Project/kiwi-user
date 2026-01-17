package wechatpay

import (
	"context"
	"net/http"

	"github.com/futurxlab/golanggraph/xerror"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WechatPayClient struct {
	AppID           string
	MchID           string
	client          *core.Client
	callbackHandler *notify.Handler
	notifyURL       string
}

func NewWechatPayClient(
	AppID string,
	MchID string,
	PrivateKeyPath string,
	PublicKeyPath string,
	PublicKeyID string,
	CertSerialNo string,
	MchAPIv3Key string,
	notifyURL string,
) (*WechatPayClient, error) {

	ctx := context.Background()

	privateKey, err := utils.LoadPrivateKeyWithPath(PrivateKeyPath)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	publicKey, err := utils.LoadPublicKeyWithPath(PublicKeyPath)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(
			MchID,
			CertSerialNo,
			privateKey,
			PublicKeyID,
			publicKey,
		),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	handler := notify.NewNotifyHandler(MchAPIv3Key, verifiers.NewSHA256WithRSAPubkeyVerifier(PublicKeyID, *publicKey))

	return &WechatPayClient{
		AppID:           AppID,
		MchID:           MchID,
		notifyURL:       notifyURL,
		client:          client,
		callbackHandler: handler,
	}, nil
}

func (w *WechatPayClient) CreatePayment(ctx context.Context, outTradeNo string, description string, openid string, currency string, amount int) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	svc := jsapi.JsapiApiService{Client: w.client}

	resp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(w.AppID),
		Mchid:       core.String(w.MchID),
		Description: core.String(description),
		OutTradeNo:  core.String(outTradeNo),
		NotifyUrl:   core.String(w.notifyURL),
		Amount: &jsapi.Amount{
			Total:    core.Int64(int64(amount)),
			Currency: core.String(currency),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(openid),
		},
	})

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return resp, nil
}

func (w *WechatPayClient) ParseRequest(ctx context.Context, req *http.Request) (*PaymentNotify, error) {
	var notify *PaymentNotify
	_, err := w.callbackHandler.ParseNotifyRequest(ctx, req, &notify)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return notify, nil
}

func (w *WechatPayClient) QueryPayment(ctx context.Context, outTradeNo string) (*payments.Transaction, error) {
	svc := jsapi.JsapiApiService{Client: w.client}

	resp, _, err := svc.QueryOrderByOutTradeNo(ctx, jsapi.QueryOrderByOutTradeNoRequest{
		Mchid:      core.String(w.MchID),
		OutTradeNo: core.String(outTradeNo),
	})

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return resp, nil
}
