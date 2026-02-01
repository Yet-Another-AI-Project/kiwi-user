package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/payment/wechatpay"
	"kiwi-user/internal/infrastructure/utils"
	"kiwi-user/internal/infrastructure/utils/aes"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/xhttp"
	"github.com/futurxlab/golanggraph/xerror"
)

type WechatPaymentService struct {
	httpClient        *xhttp.Client
	paymentRepository contract.IPaymentRepository
	userRepository    contract.IUserRepository
	AESKey            string
	ServiceNotifyURL  map[string]interface{}

	WechatPayClient wechatpay.WechatPayClient
}

type WechatPaymentStatus struct {
	TradeState     string    `json:"trade_state"`
	TradeStateDesc string    `json:"trade_state_desc"`
	SuccessTime    time.Time `json:"success_time,omitempty"`
	TransactionId  string    `json:"transaction_id,omitempty"`
}

type WechatPrepayResponse struct {
	AppID     string
	TimeStamp string
	NonceStr  string
	Package   string
	SignType  string
	PaySign   string
}

type PrepayResponse struct {
	Wechat WechatPrepayResponse
}

func NewWechatPaymentService(
	config *config.Config,
	paymentRepository contract.IPaymentRepository,
	userRepository contract.IUserRepository,
	httpClient *xhttp.Client) (*WechatPaymentService, error) {

	service := &WechatPaymentService{
		paymentRepository: paymentRepository,
		userRepository:    userRepository,
		httpClient:        httpClient,
	}

	if config.Payment != nil {
		if config.Payment.WechatAppID == "" {
			return nil, nil
		}
		client, err := wechatpay.NewWechatPayClient(
			config.Payment.WechatAppID,
			config.Payment.WechatMchID,
			config.Payment.WechatPrivateKeyPath,
			config.Payment.WechatPublicKeyPath,
			config.Payment.WechatPublicKeyID,
			config.Payment.WechatCertSerialNo,
			config.Payment.WechatMchAPIv3Key,
			config.Payment.WechatNotifyURL,
		)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		service.WechatPayClient = *client

		service.AESKey = config.Payment.AESEncryptKey
		service.ServiceNotifyURL = config.Payment.ServiceNotifyURL
	}

	return service, nil
}

func (service *WechatPaymentService) CreatePayment(ctx context.Context, encrypt string) (*entity.PaymentEntity, *PrepayResponse, error) {
	decryptData, err := aes.AESDecrypt(encrypt, []byte(service.AESKey))
	if err != nil {
		return nil, nil, xerror.Wrap(err)
	}

	var paymentRequest dto.PaymentRequestContent

	err = json.Unmarshal([]byte(decryptData), &paymentRequest)
	if err != nil {
		return nil, nil, xerror.Wrap(err)
	}

	userAggragate, err := service.userRepository.Find(ctx, paymentRequest.UserID)
	if err != nil {
		return nil, nil, xerror.Wrap(err)
	}

	if userAggragate == nil {
		return nil, nil, xerror.New("user not found")
	}

	if paymentRequest.Amount.Currency == "" {
		paymentRequest.Amount.Currency = "CNY"
	}

	payment := &entity.PaymentEntity{
		UserID: paymentRequest.UserID,
		ChannelInfo: entity.PaymentChannelInfo{
			Channel:        enum.PaymentChannel(paymentRequest.Channel),
			WechatPlatform: enum.WechatOpenIDPlatform(paymentRequest.Platform),
		},
		Service:     paymentRequest.Service,
		Amount:      paymentRequest.Amount.Total,
		Currency:    paymentRequest.Amount.Currency,
		Description: paymentRequest.Description,
	}

	outTradeNo := utils.GnerateOutTradeNo(service.WechatPayClient.MchID)

	if payment.ChannelInfo.Channel == enum.PaymentChannelWechat {
		wechatOpenID, err := service.userRepository.FindWechatOpenIDByUserAndPlatform(ctx, payment.UserID, string(payment.ChannelInfo.WechatPlatform))
		if err != nil {
			return nil, nil, xerror.Wrap(err)
		}

		if wechatOpenID == nil {
			return nil, nil, xerror.New("wechat openid not found for user")
		}

		resp, err := service.createWechatPayment(
			ctx,
			payment.Description,
			outTradeNo,
			payment.Amount,
			payment.Currency,
			wechatOpenID.OpenID,
		)
		if err != nil {
			return nil, nil, xerror.Wrap(err)
		}

		payment.OutTradeNo = outTradeNo
		payment.Status = enum.PaymentStatusNotPay
		payment.CreatedAt = time.Now()
		payment.ChannelInfo.WeChatOpenID = wechatOpenID.OpenID

		_, err = service.paymentRepository.Create(ctx, &aggregate.PaymentAggregate{
			Payment: payment,
		})
		if err != nil {
			return nil, nil, xerror.Wrap(err)
		}

		return payment, &PrepayResponse{
			Wechat: *resp,
		}, nil
	}

	return nil, nil, xerror.New("unsupported payment channel")
}

func (service *WechatPaymentService) UpdatePaymentStatus(ctx context.Context, paymentAggregate *aggregate.PaymentAggregate) (*aggregate.PaymentAggregate, error) {
	_, err := service.paymentRepository.Update(ctx, paymentAggregate)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (service *WechatPaymentService) createWechatPayment(ctx context.Context, description string, outTradeNo string, amount int, currency string, openid string) (*WechatPrepayResponse, error) {
	resp, err := service.WechatPayClient.CreatePayment(ctx, outTradeNo, description, openid, currency, amount)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &WechatPrepayResponse{
		AppID:     service.WechatPayClient.AppID,
		TimeStamp: *resp.TimeStamp,
		NonceStr:  *resp.NonceStr,
		Package:   *resp.Package,
		SignType:  *resp.SignType,
		PaySign:   *resp.PaySign,
	}, nil
}

func (service *WechatPaymentService) GetPaymentStatus(ctx context.Context, paymentAggregate *aggregate.PaymentAggregate) (*aggregate.PaymentAggregate, error) {
	transaction, err := service.WechatPayClient.QueryPayment(ctx, paymentAggregate.Payment.OutTradeNo)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	updated := false
	if transaction.TradeState != nil {
		switch *transaction.TradeState {
		case "SUCCESS":
			paymentAggregate.Payment.Status = enum.PaymentStatusSuccess
			if transaction.TransactionId != nil {
				paymentAggregate.Payment.ChannelInfo.WeChatTransactionID = *transaction.TransactionId
			}
			if transaction.SuccessTime != nil {
				successTime, parseErr := time.Parse(time.RFC3339, *transaction.SuccessTime)
				if parseErr == nil {
					paymentAggregate.Payment.PaidAt = successTime
				}
			}
			updated = true
		case "CLOSED":
			paymentAggregate.Payment.Status = enum.PaymentStatusClosed
			updated = true
		case "REFUND":
			paymentAggregate.Payment.Status = enum.PaymentStatusRefund
			updated = true
		}
	}

	if updated {
		_, err = service.paymentRepository.Update(ctx, paymentAggregate)
		if err != nil {
			return nil, xerror.Wrap(err)
		}
	}

	return paymentAggregate, nil
}

func (service *WechatPaymentService) HandlePaymentNotify(ctx context.Context, req *http.Request) (*aggregate.PaymentAggregate, error) {
	notify, err := service.WechatPayClient.ParseRequest(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	paymentAggregate, err := service.paymentRepository.FindByOutTradeNo(ctx, notify.OutTradeNo)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if paymentAggregate == nil {
		return nil, xerror.New("payment not found")
	}

	switch notify.TradeState {
	case "SUCCESS":
		paymentAggregate.Payment.Status = enum.PaymentStatusSuccess
		paymentAggregate.Payment.ChannelInfo.WeChatTransactionID = notify.TransactionID
		paymentAggregate.Payment.PaidAt = notify.SuccessTime
	case "CLOSED":
		paymentAggregate.Payment.Status = enum.PaymentStatusClosed
	case "REFUND":
		paymentAggregate.Payment.Status = enum.PaymentStatusRefund
	case "NOTPAY":
		paymentAggregate.Payment.Status = enum.PaymentStatusNotPay
	default:
		return nil, xerror.New("unsupported trade state")
	}

	_, err = service.paymentRepository.Update(ctx, paymentAggregate)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (service *WechatPaymentService) SendNotification(paymentAggregate *aggregate.PaymentAggregate) error {
	var notifyURL string
	if urlValue, ok := service.ServiceNotifyURL[string(paymentAggregate.Payment.Service)]; ok {
		notifyURL = urlValue.(string)
	} else {
		return xerror.New("service not found")
	}

	payloadData := map[string]interface{}{
		"order_no":     paymentAggregate.Payment.OutTradeNo,
		"pay_order_no": paymentAggregate.Payment.ChannelInfo.WeChatTransactionID,
		"amount":       int64(paymentAggregate.Payment.Amount),
		"pay_time":     paymentAggregate.Payment.PaidAt.Format(time.RFC3339),
		"pay_state":    string(paymentAggregate.Payment.Status),
	}

	jsonData, err := json.Marshal(payloadData)
	if err != nil {
		return xerror.Wrap(err)
	}

	encryptedData, err := aes.AESEncrypt(string(jsonData), []byte(service.AESKey))
	if err != nil {
		return xerror.Wrap(err)
	}

	requestBody := map[string]interface{}{
		"encrypt": encryptedData,
	}

	jsonReqBody, err := json.Marshal(requestBody)
	if err != nil {
		return xerror.Wrap(err)
	}

	resp, err := service.httpClient.Post(notifyURL, "application/json", bytes.NewReader(jsonReqBody))
	if err != nil {
		return xerror.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return xerror.New("notification failed")
	}

	return nil
}
