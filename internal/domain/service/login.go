package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/utils"
	"net/http"
	"net/url"
	"strings"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/alibaba/oss"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/xhttp"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"

	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type wechatUserInfo struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadimgURL string `json:"headimgurl"`
	Unionid    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMSG     string `json:"errmsg"`
}

type LoginService struct {
	logger                    logger.ILogger
	userRepository            contract.IUserRepository
	mailVertifyCodeRepository contract.IMailVertifyCodeRepository
	httpClient                *xhttp.Client
	ossClient                 *oss.AliyunOss
	config                    *config.Config

	wechatAppID                string
	wechatAppSecret            string
	wechatMiniProgramID        string
	wechatMiniProgramSecret    string
	wechatWebID                string
	wechatWebSecret            string
	wechatOfficalAccountID     string
	wechatOfficalAccountSecret string

	qyWechatCorpID     string
	qyWechatCorpSecret string

	googleClientID     string
	googleClientSecret string
}

func NewLoginService(
	logger logger.ILogger,
	config *config.Config,
	userRepository contract.IUserRepository,
	mailVertifyCodeRepository contract.IMailVertifyCodeRepository,
	httpClient *xhttp.Client,
	ossClient *oss.AliyunOss) *LoginService {

	service := &LoginService{
		logger:                    logger,
		userRepository:            userRepository,
		mailVertifyCodeRepository: mailVertifyCodeRepository,
		httpClient:                httpClient,
		ossClient:                 ossClient,
		config:                    config,
	}

	if config.Wechat != nil {
		service.wechatAppID = config.Wechat.AppID
		service.wechatAppSecret = config.Wechat.AppSecret
		service.wechatMiniProgramID = config.Wechat.MiniProgramID
		service.wechatMiniProgramSecret = config.Wechat.MiniProgramSecret
		service.wechatWebID = config.Wechat.WebID
		service.wechatWebSecret = config.Wechat.WebSecret
		service.wechatOfficalAccountID = config.Wechat.OfficalAccountID
		service.wechatOfficalAccountSecret = config.Wechat.OfficalAccountSecret
		service.qyWechatCorpID = config.Wechat.QyWechatCorpID
		service.qyWechatCorpSecret = config.Wechat.QyWechatCorpSecret
	}

	if config.Google != nil {
		service.googleClientID = config.Google.ClientID
		service.googleClientSecret = config.Google.ClientSecret
	}

	return service
}

// 微信网页登录
func (l *LoginService) WechatWebLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	refferalChannel entity.UserRefferalChannel,
	code string,
	platform string) (*aggregate.UserAggregate, error) {

	clientID := ""
	clientSecret := ""
	if platform == "officalaccount" {
		clientID = l.wechatOfficalAccountID
		clientSecret = l.wechatOfficalAccountSecret
	} else {
		clientID = l.wechatWebID
		clientSecret = l.wechatWebSecret
	}

	openid, accessToken, err := l.getWechatAccessToken(ctx, code, clientID, clientSecret)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	wechatUserInfo, err := l.getWechatUserInfo(accessToken, openid)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	l.logger.Debugf(ctx, "wechat user info: %w", wechatUserInfo)

	identity := wechatUserInfo.Unionid
	if identity == "" {
		identity = wechatUserInfo.OpenID
	}

	var userAggregate *aggregate.UserAggregate

	if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {

		var err error

		// get user by binding
		wechatBinding := &entity.BindingEntity{
			ApplicationID: application.Application.ID,
			Type:          enum.BindingTypeWechat,
			Identity:      identity,
			Verified:      true,
		}

		userAggregate, err = l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, wechatBinding)

		if err != nil {
			return xerror.Wrap(err)
		}

		if userAggregate == nil {
			// create new user
			randomUserName, err := l.randomUserName(ctx, application.Application.Name)
			if err != nil {
				return xerror.Wrap(err)
			}
			userEntity := &entity.UserEntity{
				Name:            randomUserName,
				DisplayName:     wechatUserInfo.Nickname,
				Avatar:          wechatUserInfo.HeadimgURL,
				RefferalChannel: refferalChannel,
			}

			userAggregate = &aggregate.UserAggregate{
				User:         userEntity,
				Application:  application.Application,
				Bindings:     []*entity.BindingEntity{wechatBinding},
				PersonalRole: application.DefaultPersonalRole,
			}

			userAggregate, err = l.userRepository.Create(ctx, userAggregate)

			if err != nil {
				return xerror.Wrap(err)
			}
		} else {
			// update user info
			userAggregate.User.DisplayName = wechatUserInfo.Nickname
			userAggregate.User.Avatar = wechatUserInfo.HeadimgURL
			userAggregate, err = l.userRepository.Update(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}

// 微信小程序登录
func (l *LoginService) WechatMiniprogramLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	refferalChannel entity.UserRefferalChannel,
	code string,
	miniProgramPhoneCode string) (*aggregate.UserAggregate, error) {

	var appID, appSecret string

	// 默认为小程序的配置
	appID = l.wechatMiniProgramID
	appSecret = l.wechatMiniProgramSecret

	l.logger.Infof(ctx, "start wechat mini-program login: %s, %s, %s, %s, %s", code, miniProgramPhoneCode, application.Application.Name, appID, appSecret)

	// 如果传入了 mini_program_phone_code，先通过微信接口获取手机号
	var phone string
	if miniProgramPhoneCode != "" {
		var err error
		phone, err = l.GetPhoneFromMiniProgramCode(miniProgramPhoneCode, appID, appSecret)
		if err != nil {
			l.logger.Errorf(ctx, "failed to get phone from mini program code: %w", err)
			return nil, xerror.Wrap(fmt.Errorf("failed to get phone from mini program code: %w", err))
		}
		l.logger.Infof(ctx, "got phone from mini program code: %s", phone)
	}

	_, unionId, openID, err := l.getWechatSessionKey(ctx, code, appID, appSecret)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	miniprogramOpenid := &entity.WechatOpenIDEntity{
		Platform: "miniprogram",
		OpenID:   openID,
	}

	l.logger.Infof(ctx, "wechat openid and unionid: %s, %s", openID, unionId)

	var userAggregate *aggregate.UserAggregate

	// 如果没有传入phone，保持原有逻辑
	if phone == "" {
		if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {

			var err error

			// get user by binding
			wechatBinding := &entity.BindingEntity{
				Type:          enum.BindingTypeWechat,
				Identity:      unionId,
				ApplicationID: application.Application.ID,
				Verified:      true,
			}

			userAggregate, err = l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, wechatBinding)

			if err != nil {
				return xerror.Wrap(err)
			}

			if userAggregate == nil {
				// create new user
				randomUserName, err := l.randomUserName(ctx, application.Application.Name)
				if err != nil {
					return xerror.Wrap(err)
				}

				userEntity := &entity.UserEntity{
					Name:            randomUserName,
					Avatar:          "",
					RefferalChannel: refferalChannel,
				}

				userAggregate = &aggregate.UserAggregate{
					User:          userEntity,
					Application:   application.Application,
					Bindings:      []*entity.BindingEntity{wechatBinding},
					WechatOpenIDs: []*entity.WechatOpenIDEntity{miniprogramOpenid},
					PersonalRole:  application.DefaultPersonalRole,
				}

				userAggregate, err = l.userRepository.Create(ctx, userAggregate)

				if err != nil {
					return xerror.Wrap(err)
				}
			} else {
				WechatOpenIDEntity, err := l.userRepository.FindWechatOpenIDByUserAndPlatform(ctx, userAggregate.User.ID, "miniprogram")
				if err != nil {
					return xerror.Wrap(err)
				}
				if WechatOpenIDEntity == nil {
					userAggregate.WechatOpenIDs = append(userAggregate.WechatOpenIDs, miniprogramOpenid)
					userAggregate, err = l.userRepository.Update(ctx, userAggregate)
					if err != nil {
						return xerror.Wrap(err)
					}
				}
			}

			return nil
		}); err != nil {
			return nil, xerror.Wrap(err)
		}

		return userAggregate, nil
	}

	// 传入phone时的处理逻辑
	if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		l.logger.Infof(ctx, "wechat mini program phone: %s", phone)
		var err error

		// 构建微信绑定和手机号绑定
		wechatBinding := &entity.BindingEntity{
			Type:          enum.BindingTypeWechat,
			Identity:      unionId,
			ApplicationID: application.Application.ID,
			Verified:      true,
		}

		phoneBinding := &entity.BindingEntity{
			Type:          enum.BindingTypePhone,
			Identity:      phone,
			ApplicationID: application.Application.ID,
			Verified:      true,
		}

		// 查找微信绑定对应的用户
		wechatUser, err := l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, wechatBinding)
		if err != nil {
			return xerror.Wrap(err)
		}

		// 查找手机号绑定对应的用户
		phoneUser, err := l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, phoneBinding)
		if err != nil {
			return xerror.Wrap(err)
		}

		// 情况2.1: 微信绑定存在，手机号绑定存在
		if wechatUser != nil && phoneUser != nil {
			// 如果两个绑定对应的用户不是同一个，打印warn日志
			if wechatUser.User.ID != phoneUser.User.ID {
				l.logger.Warnf(ctx, "wechat binding and phone binding belong to different users: %s, %s, %s, %s", wechatUser.User.ID, phoneUser.User.ID, unionId, phone)
			}
			// 返回微信绑定对应的用户（因为是微信登录）
			userAggregate = wechatUser
			// 确保 miniprogram openid 已添加
			WechatOpenIDEntity, err := l.userRepository.FindWechatOpenIDByUserAndPlatform(ctx, userAggregate.User.ID, "miniprogram")
			if err != nil {
				return xerror.Wrap(err)
			}
			if WechatOpenIDEntity == nil {
				userAggregate.WechatOpenIDs = append(userAggregate.WechatOpenIDs, miniprogramOpenid)
				userAggregate, err = l.userRepository.Update(ctx, userAggregate)
				if err != nil {
					return xerror.Wrap(err)
				}
			}
			return nil
		}

		// 情况2.2: 微信绑定不存在，手机号绑定存在
		if wechatUser == nil && phoneUser != nil {
			userAggregate = phoneUser
			// 检查该用户是否已有微信绑定
			hasWechatBinding := false
			for _, binding := range userAggregate.Bindings {
				if binding.Type == enum.BindingTypeWechat && binding.Identity == unionId {
					hasWechatBinding = true
					break
				}
			}
			// 创建微信绑定
			if !hasWechatBinding {
				userAggregate.Bindings = append(userAggregate.Bindings, wechatBinding)
			}
			// 确保 miniprogram openid 已添加
			WechatOpenIDEntity, err := l.userRepository.FindWechatOpenIDByUserAndPlatform(ctx, userAggregate.User.ID, "miniprogram")
			if err != nil {
				return xerror.Wrap(err)
			}
			if WechatOpenIDEntity == nil {
				userAggregate.WechatOpenIDs = append(userAggregate.WechatOpenIDs, miniprogramOpenid)
			}
			userAggregate, err = l.userRepository.Update(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
			return nil
		}

		// 情况2.3: 微信绑定存在，手机号绑定不存在
		if wechatUser != nil && phoneUser == nil {
			userAggregate = wechatUser
			// 检查该用户是否已有该手机号绑定
			hasPhoneBinding := false
			for _, binding := range userAggregate.Bindings {
				if binding.Type == enum.BindingTypePhone && binding.Identity == phone {
					hasPhoneBinding = true
					break
				}
			}
			// 创建手机号绑定
			if !hasPhoneBinding {
				userAggregate.Bindings = append(userAggregate.Bindings, phoneBinding)
			}
			// 确保 miniprogram openid 已添加
			WechatOpenIDEntity, err := l.userRepository.FindWechatOpenIDByUserAndPlatform(ctx, userAggregate.User.ID, "miniprogram")
			if err != nil {
				return xerror.Wrap(err)
			}
			if WechatOpenIDEntity == nil {
				userAggregate.WechatOpenIDs = append(userAggregate.WechatOpenIDs, miniprogramOpenid)
			}
			userAggregate, err = l.userRepository.Update(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
			return nil
		}

		// 情况2.4: 微信绑定不存在，手机号绑定不存在
		if wechatUser == nil && phoneUser == nil {
			// 创建新用户，同时创建两个绑定
			randomUserName, err := l.randomUserName(ctx, application.Application.Name)
			if err != nil {
				return xerror.Wrap(err)
			}

			userEntity := &entity.UserEntity{
				Name:            randomUserName,
				Avatar:          "",
				RefferalChannel: refferalChannel,
			}

			userAggregate = &aggregate.UserAggregate{
				User:          userEntity,
				Application:   application.Application,
				Bindings:      []*entity.BindingEntity{wechatBinding, phoneBinding},
				WechatOpenIDs: []*entity.WechatOpenIDEntity{miniprogramOpenid},
				PersonalRole:  application.DefaultPersonalRole,
			}

			userAggregate, err = l.userRepository.Create(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
			return nil
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}

func (l *LoginService) PasswordLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	name string,
	password string,
) (*aggregate.UserAggregate, error) {
	userAggregate, err := l.userRepository.FindByName(ctx, application.Application.Name, name)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if userAggregate == nil {
		return nil, xerror.Wrap(ErrUserNotFound)
	}

	if userAggregate.Application.Name != application.Application.Name {
		return nil, xerror.Wrap(ErrUserNotFound)
	}

	var passwordBinding *entity.BindingEntity
	for _, binding := range userAggregate.Bindings {
		if binding.Type == enum.BindingTypePassword {
			passwordBinding = binding
		}
	}

	if passwordBinding == nil || !passwordBinding.Verified {
		return nil, xerror.Wrap(ErrUserNotFound)
	}

	if passwordBinding.Salt == "" {
		return nil, xerror.New("salt can't be empty")
	}

	hashedPassword, err := utils.EncodePassword(password, passwordBinding.Salt)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if hashedPassword != passwordBinding.Identity {
		return nil, xerror.Wrap(ErrUserNotFound)
	}

	return userAggregate, nil
}

func (l *LoginService) getWechatSessionKey(ctx context.Context, code string, appID string, appSecret string) (sessionkey, unionid string, openid string, err error) {
	// get access token
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return "", "", "", xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", xerror.Wrap(err)
	}

	l.logger.Infof(ctx, "wechat raw response: %s", string(b))

	body := struct {
		SessionKey string `json:"session_key"`
		Unionid    string `json:"unionid"`
		OpenID     string `json:"openid"`
		Errcode    int32  `json:"errcode"`
		Errmsg     string `json:"errmsg"`
	}{}

	if err := json.Unmarshal(b, &body); err != nil {
		return "", "", "", xerror.Wrap(err)
	}

	if body.Errcode != 0 {
		// invalid code
		if body.Errcode == 40029 {
			return "", "", "", xerror.Wrap(ErrInvalidWechatCode)
		}
		return "", "", "", xerror.Wrap(fmt.Errorf("wechat api error: %d, %s", body.Errcode, body.Errmsg))
	}

	if body.Unionid == "" || body.SessionKey == "" {
		l.logger.Errorf(ctx, "missing required fields from wechat: %s, %s, %s", body.Unionid, body.SessionKey, body.OpenID)
		return "", "", "", xerror.Wrap(fmt.Errorf("session_key and unionid can't be empty"))
	}

	return body.SessionKey, body.Unionid, body.OpenID, nil
}

func (l *LoginService) getWechatAccessToken(ctx context.Context, code string, appID string, appSecret string) (string, string, error) {
	// get access token
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return "", "", xerror.Wrap(err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", xerror.Wrap(err)
	}

	body := struct {
		AccessToken string `json:"access_token"`
		Unionid     string `json:"unionid"`
		OpenID      string `json:"openid"`
		Scope       string `json:"scope"`
		Errmsg      string `json:"errmsg"`
		Errcode     int32  `json:"errcode"`
	}{}
	if err := json.Unmarshal(b, &body); err != nil {
		return "", "", xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if !strings.Contains(body.Scope, "snsapi_login") &&
			!strings.Contains(body.Scope, "snsapi_userinfo") {
			l.logger.Infof(ctx, "wechat api error: access_token scope not found or invalid: %v", body)
			return "", "", xerror.Wrap(ErrWechatInvalidScope)
		}
		return body.OpenID, body.AccessToken, nil
	}

	if body.Errcode != 0 {
		// invalid code
		if body.Errcode == 40029 {
			return "", "", xerror.Wrap(ErrInvalidWechatCode)
		}
		return "", "", xerror.Wrap(fmt.Errorf("wechat api error: %d, %s", body.Errcode, body.Errmsg))
	}

	return "", "", xerror.Wrap(errors.New("unknow server error"))
}

func (l *LoginService) getWechatUserInfo(accessToken string, openID string) (*wechatUserInfo, error) {

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		accessToken, openID)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	winfo := wechatUserInfo{}

	if err := json.Unmarshal(b, &winfo); err != nil {
		return nil, xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		return &winfo, nil
	}

	if winfo.ErrMSG != "" {
		// invalid code
		if winfo.ErrCode == 40003 {
			return nil, xerror.Wrap(ErrInvalidWechatCode)
		}
		return nil, xerror.Wrap(errors.New(winfo.ErrMSG))
	}

	return nil, xerror.Wrap(errors.New("unknow server error"))
}

func (l *LoginService) randomUserName(ctx context.Context, applicationName string) (string, error) {
	randomTokenLen := 5
	name := fmt.Sprintf("%s_%s", "user", utils.RandomToken(randomTokenLen))

	for {
		user, err := l.userRepository.FindByName(ctx, applicationName, name)
		if err != nil {
			return "", err
		}

		if user == nil {
			break
		}
		name = fmt.Sprintf("%s_%s", "user", utils.RandomToken(randomTokenLen))
	}

	return name, nil
}

func (l *LoginService) getWechatMiniProgramAccessToken(appID, appSecret string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		appID, appSecret)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	body := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	if err := json.Unmarshal(b, &body); err != nil {
		return "", xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		return body.AccessToken, nil
	}

	return "", xerror.Wrap(errors.New("unknow server error"))
}

func (l *LoginService) GetPhoneFromMiniProgramCode(code, appID, appSecret string) (string, error) {
	accessToken, err := l.getWechatMiniProgramAccessToken(appID, appSecret)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", accessToken)

	body := struct {
		Code string `json:"code"`
	}{Code: code}
	b, err := json.Marshal(body)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	resp, err := l.httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", xerror.Wrap(err)
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	defer resp.Body.Close()

	responseBody := struct {
		Errcode   int    `json:"errcode"`
		Errmsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber string `json:"phoneNumber"`
			PurePhone   string `json:"purePhoneNumber"`
			CountryCode string `json:"countryCode"`
			WaterMark   struct {
				Timestamp int64  `json:"timestamp"`
				AppID     string `json:"appid"`
			} `json:"watermark"`
		} `json:"phone_info"`
	}{}
	if err := json.Unmarshal(b, &responseBody); err != nil {
		return "", xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if responseBody.Errcode != 0 {
			// 对于 code 无效的情况，使用特定的错误类型
			if responseBody.Errcode == 40029 {
				return "", xerror.Wrap(ErrInvalidWechatCode)
			}
			return "", xerror.Wrap(fmt.Errorf("wechat api error: %d, %s", responseBody.Errcode, responseBody.Errmsg))
		}
		// 返回带区号的手机号
		return responseBody.PhoneInfo.PhoneNumber, nil
	}

	return "", xerror.Wrap(errors.New("unknow server error"))
}

// getQyWechatAccessToken 获取企业微信的 access_token
func (l *LoginService) getQyWechatAccessToken(corpID, corpSecret string) (string, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, corpSecret)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	defer resp.Body.Close()

	body := struct {
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	if err := json.Unmarshal(b, &body); err != nil {
		return "", xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if body.Errcode != 0 {
			return "", xerror.Wrap(fmt.Errorf("qy wechat api error: %d, %s", body.Errcode, body.Errmsg))
		}
		if body.AccessToken == "" {
			return "", xerror.Wrap(errors.New("access_token is empty"))
		}
		return body.AccessToken, nil
	}

	return "", xerror.Wrap(errors.New("unknown server error"))
}

type qyWechatUserInfo struct {
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
	UserID     string `json:"UserId"`
	DeviceID   string `json:"DeviceId"`
	UserTicket string `json:"user_ticket"`
	OpenID     string `json:"OpenId"`
}

type qyWechatUserDetail struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	UserID  string `json:"userid"`
	Gender  string `json:"gender"`
	Avatar  string `json:"avatar"`
	QrCode  string `json:"qr_code"`
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
	BizMail string `json:"biz_mail"`
	Address string `json:"address"`
}

// getQyWechatUserInfo 获取企业微信用户信息
func (l *LoginService) getQyWechatUserInfo(accessToken, code string) (*qyWechatUserInfo, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=%s&code=%s", accessToken, code)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	defer resp.Body.Close()

	userInfo := qyWechatUserInfo{}
	if err := json.Unmarshal(b, &userInfo); err != nil {
		return nil, xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if userInfo.Errcode != 0 {
			return nil, xerror.Wrap(fmt.Errorf("qy wechat api error: %d, %s", userInfo.Errcode, userInfo.Errmsg))
		}
		return &userInfo, nil
	}

	return nil, xerror.Wrap(errors.New("unknown server error"))
}

// getQyWechatUserDetail 获取企业微信用户敏感信息
func (l *LoginService) getQyWechatUserDetail(accessToken, userTicket string) (*qyWechatUserDetail, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserdetail?access_token=%s", accessToken)

	body := struct {
		UserTicket string `json:"user_ticket"`
	}{UserTicket: userTicket}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	resp, err := l.httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	defer resp.Body.Close()

	userDetail := qyWechatUserDetail{}
	if err := json.Unmarshal(b, &userDetail); err != nil {
		return nil, xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if userDetail.Errcode != 0 {
			return nil, xerror.Wrap(fmt.Errorf("qy wechat api error: %d, %s", userDetail.Errcode, userDetail.Errmsg))
		}
		return &userDetail, nil
	}

	return nil, xerror.Wrap(errors.New("unknown server error"))
}

type qyWechatUserProfile struct {
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
	UserID     string `json:"userid"`
	Name       string `json:"name"`
	Department []int  `json:"department"`
	Position   string `json:"position"`
	Mobile     string `json:"mobile"`
	Gender     string `json:"gender"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Status     int    `json:"status"`
}

type qyWechatDepartment struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parentid"`
	Order    int    `json:"order"`
}

type qyWechatDepartmentList struct {
	Errcode    int                  `json:"errcode"`
	Errmsg     string               `json:"errmsg"`
	Department []qyWechatDepartment `json:"department"`
}

// getQyWechatUserProfile 获取企业微信用户基本信息（包括姓名和部门）
func (l *LoginService) getQyWechatUserProfile(accessToken, userID string) (*qyWechatUserProfile, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=%s&userid=%s", accessToken, userID)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	defer resp.Body.Close()

	userProfile := qyWechatUserProfile{}
	if err := json.Unmarshal(b, &userProfile); err != nil {
		return nil, xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if userProfile.Errcode != 0 {
			return nil, xerror.Wrap(fmt.Errorf("qy wechat api error: %d, %s", userProfile.Errcode, userProfile.Errmsg))
		}
		return &userProfile, nil
	}

	return nil, xerror.Wrap(errors.New("unknown server error"))
}

// getQyWechatDepartmentName 根据部门ID获取部门名称
func (l *LoginService) getQyWechatDepartmentName(accessToken string, departmentID int) (string, error) {
	// 获取部门列表（指定 id 会返回该部门及其子部门，不指定则返回所有部门）
	// 为了准确获取单个部门信息，先尝试指定 id，如果失败则获取所有部门
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token=%s&id=%d", accessToken, departmentID)

	resp, err := l.httpClient.Get(url)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", xerror.Wrap(err)
	}

	defer resp.Body.Close()

	deptList := qyWechatDepartmentList{}
	if err := json.Unmarshal(b, &deptList); err != nil {
		return "", xerror.Wrap(err)
	}

	if resp.StatusCode == http.StatusOK {
		if deptList.Errcode != 0 {
			return "", xerror.Wrap(fmt.Errorf("qy wechat api error: %d, %s", deptList.Errcode, deptList.Errmsg))
		}
		// 查找匹配的部门
		for _, dept := range deptList.Department {
			if dept.ID == departmentID {
				return dept.Name, nil
			}
		}
		return "", xerror.Wrap(errors.New("department not found"))
	}

	return "", xerror.Wrap(errors.New("unknown server error"))
}

// QyWechatLogin 企业微信登录
func (l *LoginService) QyWechatLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	refferalChannel entity.UserRefferalChannel,
	code string,
) (*aggregate.UserAggregate, error) {
	l.logger.Infof(ctx, "start qy wechat login: %s, %s", code, application.Application.Name)

	// 1. 获取企业微信 access_token
	accessToken, err := l.getQyWechatAccessToken(l.qyWechatCorpID, l.qyWechatCorpSecret)
	if err != nil {
		l.logger.Errorf(ctx, "failed to get qy wechat access token: %w", err)
		return nil, xerror.Wrap(err)
	}

	// 2. 获取企业微信用户信息
	userInfo, err := l.getQyWechatUserInfo(accessToken, code)
	if err != nil {
		l.logger.Errorf(ctx, "failed to get qy wechat user info: %w", err)
		return nil, xerror.Wrap(err)
	}

	// 企业微信接口返回：如果是企业员工返回 UserID，非企业员工返回 OpenID
	if userInfo.UserID == "" && userInfo.OpenID == "" {
		return nil, xerror.Wrap(errors.New("both userID and openID are empty"))
	}

	l.logger.Infof(ctx, "qy wechat user info: %s, %s, %s", userInfo.UserID, userInfo.OpenID, userInfo.UserTicket)

	// 3. 获取用户详细信息（包括 avatar）
	var avatar string
	if userInfo.UserTicket != "" {
		userDetail, err := l.getQyWechatUserDetail(accessToken, userInfo.UserTicket)
		if err != nil {
			l.logger.Warnf(ctx, "failed to get qy wechat user detail: %w", err)
			// 如果获取详细信息失败，继续使用基本信息登录
		} else {
			avatar = userDetail.Avatar
			l.logger.Infof(ctx, "qy wechat user detail: %s", avatar)
		}
	}

	// 4. 获取用户真实姓名和部门（仅企业员工，有 UserID）
	var realName string
	var department string
	if userInfo.UserID != "" {
		userProfile, err := l.getQyWechatUserProfile(accessToken, userInfo.UserID)
		if err != nil {
			l.logger.Warnf(ctx, "failed to get qy wechat user profile: %w", err)
			// 获取失败不影响登录，只打印日志
		} else {
			realName = userProfile.Name
			l.logger.Infof(ctx, "qy wechat user profile: %s, %v", realName, userProfile.Department)

			// 获取部门名称（取第一个部门）
			if len(userProfile.Department) > 0 {
				deptName, err := l.getQyWechatDepartmentName(accessToken, userProfile.Department[0])
				if err != nil {
					l.logger.Warnf(ctx, "failed to get qy wechat department name: %w", err)
					// 获取部门名称失败不影响登录，只打印日志
				} else {
					department = deptName
					l.logger.Infof(ctx, "qy wechat department name: %s", department)
				}
			}
		}
	}

	var userAggregate *aggregate.UserAggregate

	// 4. 查找或创建用户（统一处理企业员工和非企业员工）
	// 区别在于：企业员工 qy_wechat_user_id 字段有值，非企业员工 open_id 字段有值
	// 所有用户都需要创建 qy_wechat binding，binding 的 identity 使用 userid（企业员工）或 openid（非企业员工）

	// 确定用于查找和 binding 的 identity
	bindingIdentity := userInfo.UserID
	if bindingIdentity == "" {
		bindingIdentity = userInfo.OpenID
	}

	if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		// 先通过 binding 查找用户（统一通过 qy_wechat binding 查找）
		qyWechatBinding := &entity.BindingEntity{
			ApplicationID: application.Application.ID,
			Type:          enum.BindingTypeQyWechat,
			Identity:      bindingIdentity,
			Verified:      true,
		}

		userAggregate, err = l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, qyWechatBinding)
		if err != nil {
			return xerror.Wrap(err)
		}

		if userAggregate == nil {
			// 创建新用户
			// 如果获取到了真实姓名，使用真实姓名；否则使用随机用户名
			var userName string
			if realName != "" {
				userName = realName
			} else {
				randomUserName, err := l.randomUserName(ctx, application.Application.Name)
				if err != nil {
					return xerror.Wrap(err)
				}
				userName = randomUserName
			}

			userEntity := &entity.UserEntity{
				Name:            userName,
				Avatar:          avatar,
				RefferalChannel: refferalChannel,
				Department:      department, // 如果获取到了部门，使用部门；否则为空字符串
			}

			// 创建 qy_wechat_user_id 实体
			// 企业员工：qy_wechat_user_id 字段有值（userid），open_id 字段可选
			// 非企业员工：qy_wechat_user_id 字段为空，open_id 字段有值（openid）
			qyWechatUserID := &entity.QyWechatUserIDEntity{
				QyWechatUserID: userInfo.UserID, // 企业员工有值，非企业员工为空
				OpenID:         userInfo.OpenID,
			}

			userAggregate = &aggregate.UserAggregate{
				User:            userEntity,
				Application:     application.Application,
				Bindings:        []*entity.BindingEntity{qyWechatBinding},
				QyWechatUserIDs: []*entity.QyWechatUserIDEntity{qyWechatUserID},
				PersonalRole:    application.DefaultPersonalRole,
			}

			userAggregate, err = l.userRepository.Create(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
		} else {
			// 用户已存在，更新用户信息和 qy_wechat_user_ids
			userAggregate.User.Avatar = avatar

			// 检查是否已有对应的 qy_wechat_user_id 记录
			hasQyWechatUserID := false
			if userInfo.UserID != "" {
				// 企业员工：通过 qy_wechat_user_id 字段匹配
				for _, qyWechatUserID := range userAggregate.QyWechatUserIDs {
					if qyWechatUserID.QyWechatUserID == userInfo.UserID {
						// 更新 open_id（如果有）
						if userInfo.OpenID != "" {
							qyWechatUserID.OpenID = userInfo.OpenID
						}
						hasQyWechatUserID = true
						break
					}
				}
			} else {
				// 非企业员工：通过 open_id 字段匹配（且 qy_wechat_user_id 为空）
				for _, qyWechatUserID := range userAggregate.QyWechatUserIDs {
					if qyWechatUserID.OpenID == userInfo.OpenID && qyWechatUserID.QyWechatUserID == "" {
						hasQyWechatUserID = true
						break
					}
				}
			}

			if !hasQyWechatUserID {
				// 添加新的 qy_wechat_user_id 记录
				qyWechatUserID := &entity.QyWechatUserIDEntity{
					QyWechatUserID: userInfo.UserID,
					OpenID:         userInfo.OpenID,
				}
				userAggregate.QyWechatUserIDs = append(userAggregate.QyWechatUserIDs, qyWechatUserID)
			}

			// 检查是否已有 qy_wechat binding（所有用户都需要）
			hasQyWechatBinding := false
			for _, binding := range userAggregate.Bindings {
				if binding.Type == enum.BindingTypeQyWechat && binding.Identity == bindingIdentity {
					hasQyWechatBinding = true
					break
				}
			}
			if !hasQyWechatBinding {
				qyWechatBinding.ID = uuid.Nil
				userAggregate.Bindings = append(userAggregate.Bindings, qyWechatBinding)
			}

			userAggregate, err = l.userRepository.Update(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}

func (l *LoginService) PhoneLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	phone string,
) (*aggregate.UserAggregate, error) {
	var userAggregate *aggregate.UserAggregate

	if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		// 1. 通过手机号查找用户
		phoneBinding := &entity.BindingEntity{
			Type:          enum.BindingTypePhone,
			Identity:      phone,
			ApplicationID: application.Application.ID,
			Verified:      true,
		}

		userAggregate, err = l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, phoneBinding)
		if err != nil {
			l.logger.Debugf(ctx, "phone login FindByBindingForUpdate error: %w", err)
			return xerror.Wrap(err)
		}

		// 2. 如果用户不存在，创建新用户
		if userAggregate == nil {
			l.logger.Debugf(ctx, "phone login create user start: %w", err)
			// 生成随机用户名
			randomUserName, err := l.randomUserName(ctx, application.Application.Name)
			if err != nil {
				return xerror.Wrap(err)
			}

			// 创建用户实体
			userEntity := &entity.UserEntity{
				Name:        randomUserName,
				DisplayName: randomUserName,
			}

			// 创建用户聚合
			userAggregate = &aggregate.UserAggregate{
				User:         userEntity,
				Application:  application.Application,
				Bindings:     []*entity.BindingEntity{phoneBinding},
				PersonalRole: application.DefaultPersonalRole,
			}

			// 保存到数据库
			userAggregate, err = l.userRepository.Create(ctx, userAggregate)
			if err != nil {
				l.logger.Debugf(ctx, "phone login create user error: %w", err)
				return xerror.Wrap(err)
			}
			l.logger.Debugf(ctx, "phone login create user success: %w", err)
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}

// EmailLogin handles email verification code login
func (l *LoginService) EmailLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	email string,
) (*aggregate.UserAggregate, error) {
	var userAggregate *aggregate.UserAggregate

	if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		// 1. Find user by email binding
		emailBinding := &entity.BindingEntity{
			Type:          enum.BindingTypeEmail,
			Identity:      email,
			ApplicationID: application.Application.ID,
			Verified:      true,
		}

		userAggregate, err = l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, emailBinding)
		if err != nil {
			return xerror.Wrap(err)
		}

		// 2. If user doesn't exist, create new user
		if userAggregate == nil {
			// Generate random username
			randomUserName, err := l.randomUserName(ctx, application.Application.Name)
			if err != nil {
				return xerror.Wrap(err)
			}

			// Create user entity
			userEntity := &entity.UserEntity{
				Name:        randomUserName,
				DisplayName: randomUserName,
			}

			// Create user aggregate
			userAggregate = &aggregate.UserAggregate{
				User:         userEntity,
				Application:  application.Application,
				Bindings:     []*entity.BindingEntity{emailBinding},
				PersonalRole: application.DefaultPersonalRole,
			}

			// Save to database
			userAggregate, err = l.userRepository.Create(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}

type googleUserInfo struct {
	Sub           string // Google user ID (unique identifier)
	Email         string // User's email address
	EmailVerified bool   // Whether the email is verified
	Name          string // User's full name
	GivenName     string // User's first name
	FamilyName    string // User's last name
	Picture       string // URL of user's profile picture
	Locale        string // User's locale
}

type googleTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// getGoogleIdTokenFromCode 使用授权码换取 Google ID token
func (l *LoginService) getGoogleIdTokenFromCode(ctx context.Context, code string, redirectURI string) (string, error) {
	idToken, _, err := l.getGoogleTokensFromCode(ctx, code, redirectURI)
	return idToken, err
}

// getGoogleTokensFromCode 使用授权码换取 Google ID token 和 access token
func (l *LoginService) getGoogleTokensFromCode(ctx context.Context, code string, redirectURI string) (string, string, error) {
	if l.googleClientID == "" || l.googleClientSecret == "" {
		return "", "", xerror.Wrap(errors.New("google client credentials are not configured"))
	}

	// 构建请求参数
	apiURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", l.googleClientID)
	data.Set("client_secret", l.googleClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := l.httpClient.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", xerror.Wrap(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", xerror.Wrap(err)
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(b, &tokenResp); err != nil {
		return "", "", xerror.Wrap(err)
	}

	if tokenResp.Error != "" {
		return "", "", xerror.Wrap(fmt.Errorf("failed to exchange code for token: %s - %s", tokenResp.Error, tokenResp.ErrorDescription))
	}

	if tokenResp.IDToken == "" {
		return "", "", xerror.Wrap(errors.New("id_token is empty in response"))
	}

	if tokenResp.AccessToken == "" {
		return "", "", xerror.Wrap(errors.New("access_token is empty in response"))
	}

	return tokenResp.IDToken, tokenResp.AccessToken, nil
}

// verifyGoogleIdToken 使用 Google 官方 SDK 验证 ID token
// 使用 google.golang.org/api/idtoken 包，这是生产环境推荐方案
func (l *LoginService) verifyGoogleIdToken(ctx context.Context, idToken string) (*googleUserInfo, error) {
	// 检查是否配置了 Client ID
	if l.googleClientID == "" {
		return nil, xerror.Wrap(errors.New("google client ID is not configured"))
	}

	// 使用官方 SDK 验证 ID token
	// Validate 会自动验证签名、iss、aud、exp 等字段
	payload, err := idtoken.Validate(ctx, idToken, l.googleClientID)
	if err != nil {
		l.logger.Errorf(ctx, "failed to validate google id token: %w", err)
		return nil, xerror.Wrap(fmt.Errorf("invalid id token: %w", err))
	}

	l.logger.Debugf(ctx, "google id token validated: sub=%s, iss=%s, aud=%s", payload.Subject, payload.Issuer, payload.Audience)

	// 验证 issuer 是否为 Google
	if payload.Issuer != "https://accounts.google.com" && payload.Issuer != "accounts.google.com" {
		return nil, xerror.Wrap(fmt.Errorf("invalid token issuer: %s", payload.Issuer))
	}

	// 构建用户信息
	userInfo := &googleUserInfo{
		Sub: payload.Subject,
	}

	// 从 Claims 中提取可选字段（需要用户授权了 profile 和 email scope）
	if email, ok := payload.Claims["email"].(string); ok {
		userInfo.Email = email
	}
	if emailVerified, ok := payload.Claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}
	if name, ok := payload.Claims["name"].(string); ok {
		userInfo.Name = name
	}
	if givenName, ok := payload.Claims["given_name"].(string); ok {
		userInfo.GivenName = givenName
	}
	if familyName, ok := payload.Claims["family_name"].(string); ok {
		userInfo.FamilyName = familyName
	}
	if picture, ok := payload.Claims["picture"].(string); ok {
		userInfo.Picture = picture
	}
	if locale, ok := payload.Claims["locale"].(string); ok {
		userInfo.Locale = locale
	}

	return userInfo, nil
}

// getGoogleUserInfoFromAccessToken 使用 access token 调用 userinfo API 获取用户信息
func (l *LoginService) getGoogleUserInfoFromAccessToken(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	// 使用 OpenID Connect userinfo endpoint
	apiURL := "https://openidconnect.googleapis.com/v1/userinfo"

	// 使用标准 http.Client 发送带 Authorization header 的请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// 使用标准 http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, xerror.Wrap(fmt.Errorf("failed to get userinfo: status %d, body: %s", resp.StatusCode, string(body)))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var userInfo googleUserInfo
	if err := json.Unmarshal(b, &userInfo); err != nil {
		return nil, xerror.Wrap(err)
	}

	return &userInfo, nil
}

// uploadGooglePictureToOSS 下载 Google 头像并上传到 OSS
func (l *LoginService) uploadGooglePictureToOSS(ctx context.Context, pictureURL string, userID string) (string, error) {
	if pictureURL == "" {
		return "", nil
	}

	l.logger.Infof(ctx, "uploading google picture to OSS: %s", pictureURL)

	if l.ossClient == nil || l.config == nil || l.config.OSS == nil {
		l.logger.Warnf(ctx, "OSS client or config not available, skipping picture upload")
		return pictureURL, nil
	}

	// 下载图片
	resp, err := l.httpClient.Get(pictureURL)
	if err != nil {
		return "", xerror.Wrap(fmt.Errorf("failed to download picture: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", xerror.Wrap(fmt.Errorf("failed to download picture: status %d", resp.StatusCode))
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", xerror.Wrap(fmt.Errorf("failed to read picture data: %w", err))
	}

	// 上传到 OSS
	key := fmt.Sprintf("user/user_avatar/%s", userID)
	if err := l.ossClient.PutObject(l.config.OSS.BucketName, key, bytes.NewReader(imageData)); err != nil {
		return "", xerror.Wrap(fmt.Errorf("failed to upload picture to OSS: %w", err))
	}

	// 返回 CDN URL
	if l.config.OSS.CDN != "" {
		return fmt.Sprintf("https://%s/%s", l.config.OSS.CDN, key), nil
	}

	// 如果没有 CDN，返回 OSS URL
	return fmt.Sprintf("https://%s.%s/%s", l.config.OSS.BucketName, l.config.OSS.Endpoint, key), nil
}

// GoogleWebLogin Google Web 登录
func (l *LoginService) GoogleWebLogin(
	ctx context.Context,
	application *aggregate.ApplicationAggregate,
	refferalChannel entity.UserRefferalChannel,
	code string,
	redirectURI string,
) (*aggregate.UserAggregate, error) {
	l.logger.Infof(ctx, "start google web login: %s", application.Application.Name)

	// 1. 使用授权码换取 ID token 和 access token
	idToken, _, err := l.getGoogleTokensFromCode(ctx, code, redirectURI)
	if err != nil {
		l.logger.Errorf(ctx, "failed to exchange code for tokens: %w", err)
		return nil, xerror.Wrap(err)
	}

	// 2. 使用 Google 官方 SDK 验证 ID token 并获取 Google 用户信息
	googleUserInfo, err := l.verifyGoogleIdToken(ctx, idToken)
	if err != nil {
		l.logger.Errorf(ctx, "failed to verify google id token: %w", err)
		return nil, xerror.Wrap(err)
	}

	l.logger.Infof(ctx, "google user info: sub=%s, email=%s, name=%s", googleUserInfo.Sub, googleUserInfo.Email, googleUserInfo.Name)

	// 使用 sub (subject) 作为用户的唯一标识
	identity := googleUserInfo.Sub

	var userAggregate *aggregate.UserAggregate

	if err := l.userRepository.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		// 2. 通过 Google binding 查找用户
		googleBinding := &entity.BindingEntity{
			ApplicationID: application.Application.ID,
			Type:          enum.BindingTypeGoogle,
			Identity:      identity,
			Verified:      true,
		}

		userAggregate, err = l.userRepository.FindByBindingForUpdate(ctx, application.Application.ID, googleBinding)

		if err != nil {
			return xerror.Wrap(err)
		}

		if userAggregate == nil {
			// 3. 创建新用户
			randomUserName, err := l.randomUserName(ctx, application.Application.Name)
			if err != nil {
				return xerror.Wrap(err)
			}

			// 使用 Google 提供的名称，如果没有则使用随机用户名
			displayName := googleUserInfo.Name
			if displayName == "" {
				displayName = randomUserName
			}

			// 先创建用户以获取 userID，然后再上传头像
			// 这里先使用原始 URL，上传会在创建后完成
			userEntity := &entity.UserEntity{
				Name:            randomUserName,
				DisplayName:     displayName,
				Avatar:          googleUserInfo.Picture,
				RefferalChannel: refferalChannel,
			}

			userAggregate = &aggregate.UserAggregate{
				User:         userEntity,
				Application:  application.Application,
				Bindings:     []*entity.BindingEntity{googleBinding},
				PersonalRole: application.DefaultPersonalRole,
			}

			userAggregate, err = l.userRepository.Create(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}

			// 上传头像到 OSS
			if googleUserInfo.Picture != "" {
				ossAvatarURL, err := l.uploadGooglePictureToOSS(ctx, googleUserInfo.Picture, userAggregate.User.ID)
				if err != nil {
					l.logger.Warnf(ctx, "failed to upload avatar to OSS: %v, using original URL", err)
					// 如果上传失败，继续使用原始URL
				} else if ossAvatarURL != "" {
					// 更新头像为 OSS URL
					userAggregate.User.Avatar = ossAvatarURL
					userAggregate, err = l.userRepository.Update(ctx, userAggregate)
					if err != nil {
						l.logger.Warnf(ctx, "failed to update user avatar: %v", err)
						// 不返回错误，因为用户已经创建成功
					}
				}
			}

		} else {
			// 4. 用户已存在，更新用户信息
			if googleUserInfo.Name != "" {
				userAggregate.User.DisplayName = googleUserInfo.Name
			}
			if googleUserInfo.Picture != "" {
				// 上传头像到 OSS
				ossAvatarURL, err := l.uploadGooglePictureToOSS(ctx, googleUserInfo.Picture, userAggregate.User.ID)
				if err != nil {
					l.logger.Warnf(ctx, "failed to upload avatar to OSS: %v, using original URL", err)
					// 如果上传失败，使用原始URL
					userAggregate.User.Avatar = googleUserInfo.Picture
				} else if ossAvatarURL != "" {
					userAggregate.User.Avatar = ossAvatarURL
				} else {
					userAggregate.User.Avatar = googleUserInfo.Picture
				}
			}
			userAggregate, err = l.userRepository.Update(ctx, userAggregate)
			if err != nil {
				return xerror.Wrap(err)
			}
		}

		return nil
	}); err != nil {
		return nil, xerror.Wrap(err)
	}

	return userAggregate, nil
}
