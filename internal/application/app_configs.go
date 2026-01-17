package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kiwi-user/config"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/cache"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var jsAPIListMapping = map[string][]string{
	"wx_scan":          {"scanQRCode"},
	"wx_downloadimage": {"previewImage", "uploadImage", "downloadImage"},
}

// 微信配置缓存项
type wxConfigCacheItem struct {
	Config    *dto.WxAppConfigResponse
	ExpiresAt time.Time
}

type ConfigApplication struct {
	config   *config.Config
	logger   logger.ILogger
	memCache *cache.MemCache
	redis    *redis.Client
	sf       singleflight.Group
}

func NewConfigApplication(
	config *config.Config,
	logger logger.ILogger,
	memCache *cache.MemCache,
	redisClient *redis.Client,
) *ConfigApplication {
	return &ConfigApplication{
		config:   config,
		logger:   logger,
		memCache: memCache,
		redis:    redisClient,
	}
}

func (c *ConfigApplication) GetWxOpenConfig(ctx context.Context, req *dto.WxAppConfigRequest) (*dto.WxAppConfigResponse, *facade.Error) {
	jsAPIList, ok := jsAPIListMapping[req.Scene]
	if !ok {
		return nil, facade.ErrBadRequest.Wrap(errors.New("未找到配置"))
	}

	if c.config == nil || c.config.Wechat == nil {
		return nil, facade.ErrBadRequest.Wrap(errors.New("未找到配置"))
	}

	if req.URL == "" {
		return nil, facade.ErrBadRequest.Wrap(errors.New("url 不能为空"))
	}
	if wxCon, err := c.GetWxOpenConfigByCache(ctx, req); err == nil && wxCon != nil {
		c.logger.Infof(ctx, "GetWxOpenConfigByCache success %v", wxCon)
		return wxCon, nil
	}

	// 使用 singleflight 防止缓存击穿
	cacheKey := fmt.Sprintf("wx_config_%s", req.URL)
	v, err, _ := c.sf.Do(cacheKey, func() (interface{}, error) {
		return c.GetWechatConfig(ctx, c.config.Wechat.OfficalAccountID, c.config.Wechat.OfficalAccountSecret, req.URL, jsAPIList)
	})
	if err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	if wxCon, ok := v.(*dto.WxAppConfigResponse); ok {
		return wxCon, nil
	}

	return nil, facade.ErrServerInternal.Facade("failed to get wechat config")
}

func (c *ConfigApplication) GetWxOpenConfigByCache(ctx context.Context, req *dto.WxAppConfigRequest) (*dto.WxAppConfigResponse, *facade.Error) {
	// 生成缓存键：使用URL作为唯一标识
	cacheKey := fmt.Sprintf("wx_config_%s", req.URL)

	// 从Redis获取缓存数据
	cachedData, err := c.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err != redis.Nil {
			c.logger.Errorf(ctx, "从Redis获取微信配置失败: %w", err)
		}
		return nil, nil
	}

	// 反序列化缓存数据
	var item wxConfigCacheItem
	if err := json.Unmarshal([]byte(cachedData), &item); err != nil {
		c.logger.Errorf(ctx, "反序列化微信配置缓存数据失败: %w", err)
		return nil, nil
	}

	// 检查是否过期
	if time.Now().Before(item.ExpiresAt) {
		c.logger.Infof(ctx, "从缓存获取微信配置成功 %s", req.URL)
		return item.Config, nil
	} else {
		// 已过期，删除缓存
		err := c.redis.Del(ctx, cacheKey).Err()
		if err != nil {
			c.logger.Errorf(ctx, "从Redis删除过期的微信配置缓存失败: %w", err)
		}
		c.logger.Infof(ctx, "微信配置缓存已过期 %s", req.URL)
	}

	return nil, nil
}

func (c *ConfigApplication) SetWxOpenConfigByCache(url string, expiresIn float64, wechatConfig *dto.WxAppConfigResponse) {
	// 把wechatConfig保存在Redis缓存，过期时间为 expiresIn 单位s
	cacheKey := fmt.Sprintf("wx_config_%s", url)
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	cacheItem := &wxConfigCacheItem{
		Config:    wechatConfig,
		ExpiresAt: expiresAt,
	}

	// 序列化缓存数据
	data, err := json.Marshal(cacheItem)
	if err != nil {
		c.logger.Errorf(context.Background(), "序列化微信配置缓存数据失败: %w", err)
		return
	}

	// 存储到Redis，设置过期时间
	ttl := time.Duration(expiresIn) * time.Second
	err = c.redis.Set(context.Background(), cacheKey, data, ttl).Err()
	if err != nil {
		c.logger.Errorf(context.Background(), "设置微信配置到Redis失败: %w", err)
	}
}

func (c *ConfigApplication) GetWechatConfig(ctx context.Context, appid, secret, url string, jsAPIList []string) (*dto.WxAppConfigResponse, error) {
	// 获取 access_token
	tokenEndpoint := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret
	tokenResponse, err := http.Get(tokenEndpoint)
	if err != nil {
		return &dto.WxAppConfigResponse{}, err
	}
	defer tokenResponse.Body.Close()

	tokenData, err := io.ReadAll(tokenResponse.Body)
	if err != nil {
		return &dto.WxAppConfigResponse{}, err
	}

	var tokenResult map[string]interface{}
	if err = json.Unmarshal(tokenData, &tokenResult); err != nil {
		return nil, err
	}
	if errMsg, ok := tokenResult["errmsg"]; ok && errMsg != "" {
		return nil, fmt.Errorf("获取 access_token 失败: %s", errMsg)
	}
	accessToken := tokenResult["access_token"].(string)

	// 获取 jsapi_ticket
	ticketEndpoint := "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=" + accessToken + "&type=jsapi"
	ticketResponse, err := http.Get(ticketEndpoint)
	if err != nil {
		return &dto.WxAppConfigResponse{}, err
	}
	defer ticketResponse.Body.Close()

	ticketData, err := io.ReadAll(ticketResponse.Body)
	if err != nil {
		return &dto.WxAppConfigResponse{}, err
	}

	var ticketResult map[string]interface{}
	if err = json.Unmarshal(ticketData, &ticketResult); err != nil {
		return nil, err
	}
	var jsapiTicket string
	var expiresIn float64
	c.logger.Infof(ctx, "ticketData %v", ticketData)
	if ticket, ok := ticketResult["ticket"]; ok {
		jsapiTicket = ticket.(string)
	}
	if _, ok := ticketResult["expires_in"]; ok {
		expiresIn = ticketResult["expires_in"].(float64)
	}
	// 生成随机字符串
	nonceStr := utils.RandomString(16)

	// 创建时间戳
	timestamp := time.Now().Unix()

	// 生成签名
	signatureString := "jsapi_ticket=" + jsapiTicket + "&noncestr=" + nonceStr + "&timestamp=" + strconv.FormatInt(timestamp, 10) + "&url=" + url
	signature := utils.Sha1(signatureString)

	wechatConfig := &dto.WxAppConfigResponse{
		AppId:     appid,
		Timestamp: timestamp,
		NonceStr:  nonceStr,
		Signature: signature,
		JsAPIList: jsAPIList,
		Debug:     true,
	}

	// 存到缓存中
	if expiresIn > 0 {
		c.SetWxOpenConfigByCache(url, expiresIn, wechatConfig)
		c.logger.Infof(ctx, "设置微信配置缓存成功")
	}
	return wechatConfig, nil
}
