package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kiwi-user/config"
	"kiwi-user/internal/facade/dto"
	"net/http"
	"strings"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/cache"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
)

type mediaTokenCacheItem struct {
	token     string
	expiresAt time.Time
}

type MediaApplication struct {
	logger          logger.ILogger
	wechatAppID     string
	wechatAppSecret string
	memCache        *cache.MemCache
	sf              singleflight.Group
	redis           *redis.Client
}

func NewMediaApplication(config *config.Config, futurxLogger logger.ILogger, memCache *cache.MemCache, redisClient *redis.Client) *MediaApplication {
	return &MediaApplication{
		wechatAppID:     config.Wechat.OfficalAccountID,
		wechatAppSecret: config.Wechat.OfficalAccountSecret,
		logger:          futurxLogger,
		memCache:        memCache,
		redis:           redisClient,
	}
}

func (a *MediaApplication) GetMediaToken(ctx context.Context) (accessToken string) {
	// 优先从缓存获取
	if accessToken = a.GetTokenByCache(ctx); accessToken != "" {
		a.logger.Infof(ctx, "从缓存获取MediaToken")
		return
	}

	// 使用 singleflight 防止缓存击穿
	key := a.getMediaTokenCacheKey()
	v, err, _ := a.sf.Do(key, func() (interface{}, error) {
		a.logger.Infof(ctx, "通过 singleflight 获取 MediaToken")
		return a.fetchAndCacheToken(ctx)
	})

	if err != nil {
		a.logger.Errorf(ctx, "通过 singleflight 获取 MediaToken 失败: %w", err)
		return ""
	}

	if token, ok := v.(string); ok {
		return token
	}

	return ""
}

func (a *MediaApplication) fetchAndCacheToken(ctx context.Context) (string, error) {
	tokenEndpoint := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + a.wechatAppID + "&secret=" + a.wechatAppSecret
	tokenResponse, err := http.Get(tokenEndpoint)
	if err != nil {
		a.logger.Errorf(ctx, "GetMediaToken: %w", err)
		return "", err
	}
	defer tokenResponse.Body.Close()

	tokenData, err := io.ReadAll(tokenResponse.Body)
	if err != nil {
		a.logger.Errorf(ctx, "GetMediaToken Read Error: %w", err)
		return "", err
	}
	tokenResp := map[string]interface{}{}

	if err = json.Unmarshal(tokenData, &tokenResp); err != nil {
		a.logger.Errorf(ctx, "GetMediaToken Unmarshal Error: %w", err)
		return "", err
	}

	if errmsg, ok := tokenResp["errmsg"]; ok && errmsg != "" {
		return "", errors.New(errmsg.(string))
	}
	var accessToken string
	if accessTokenInter, ok := tokenResp["access_token"]; ok {
		accessToken = accessTokenInter.(string)
	}

	expiresIn := 0.0
	if v, ok := tokenResp["expires_in"]; ok {
		switch vv := v.(type) {
		case float64:
			expiresIn = vv
		case int:
			expiresIn = float64(vv)
		}
	}

	if expiresIn-60 > 0 {
		a.SetAccessTokenIntoCache(accessToken, expiresIn-60)
		a.logger.Infof(ctx, "mediaToken 写入缓存")
	}

	return accessToken, nil
}

func (a *MediaApplication) getMediaTokenCacheKey() string {
	cacheKey := fmt.Sprintf("media_token_%s", a.wechatAppID)
	return cacheKey
}

func (a *MediaApplication) GetTokenByCache(ctx context.Context) string {
	cacheKey := a.getMediaTokenCacheKey()

	// 从Redis获取缓存数据
	cachedData, err := a.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if err != redis.Nil {
			a.logger.Errorf(ctx, "从Redis获取MediaToken失败: %w", err)
		}
		return ""
	}

	// 反序列化缓存数据
	var item mediaTokenCacheItem
	if err = json.Unmarshal([]byte(cachedData), &item); err != nil {
		a.logger.Errorf(ctx, "反序列化MediaToken缓存数据失败: %w", err)
		return ""
	}

	// 检查token是否有效且未过期
	if item.token != "" && time.Now().Before(item.expiresAt) {
		return item.token
	}

	return ""
}

func (a *MediaApplication) SetAccessTokenIntoCache(accessToken string, expiresIn float64) {
	if accessToken != "" && expiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
		item := &mediaTokenCacheItem{
			token:     accessToken,
			expiresAt: expiresAt,
		}

		// 序列化缓存数据
		data, err := json.Marshal(item)
		if err != nil {
			a.logger.Errorf(context.Background(), "序列化MediaToken缓存数据失败: %w", err)
			return
		}

		// 存储到Redis，设置过期时间
		cacheKey := a.getMediaTokenCacheKey()
		ttl := time.Duration(expiresIn) * time.Second
		err = a.redis.Set(context.Background(), cacheKey, data, ttl).Err()
		if err != nil {
			a.logger.Errorf(context.Background(), "设置MediaToken到Redis失败: %w", err)
		}
	}
}

func (a *MediaApplication) GetWechatMedia(ctx context.Context, resourceID string) (*dto.WechatMediaResponse, *facade.Error) {
	return a.getWechatMediaRecursive(ctx, resourceID, 0)
}

func (a *MediaApplication) getWechatMediaRecursive(ctx context.Context, resourceID string, retryCount int) (*dto.WechatMediaResponse, *facade.Error) {
	if retryCount >= 3 {
		return nil, facade.ErrBadRequest.Facade("获取微信媒体资源失败，已重试3次")
	}

	accessToken := a.GetMediaToken(ctx)
	if accessToken == "" {
		return nil, facade.ErrBadRequest.Facade("获取token失败")
	}
	mediaEndpoint := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s", accessToken, resourceID)
	mediaResponse, err := http.Get(mediaEndpoint)
	if err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}
	defer mediaResponse.Body.Close()

	// 检查响应头，判断媒体类型
	contentType := mediaResponse.Header.Get("Content-Type")
	a.logger.Infof(ctx, "获取音频结果: %s", contentType)

	// 如果微信返回JSON，说明是错误信息
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/plain") {
		apiErr := map[string]interface{}{}
		if err = json.NewDecoder(mediaResponse.Body).Decode(&apiErr); err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}
		a.logger.Errorf(ctx, "获取音频结果失败: %w", apiErr)

		if errMsg, ok := apiErr["errmsg"]; ok {
			if errMsgStr, ok := errMsg.(string); ok {
				// 当 access_token 失效时，删除缓存并重试
				if strings.Contains(errMsgStr, "access_token is invalid") {
					a.logger.Warnf(ctx, "access_token 无效，准备重试...: %d", retryCount+1)
					// 从Redis删除缓存
					err := a.redis.Del(ctx, a.getMediaTokenCacheKey()).Err()
					if err != nil {
						a.logger.Errorf(ctx, "从Redis删除MediaToken缓存失败: %w", err)
					}
					return a.getWechatMediaRecursive(ctx, resourceID, retryCount+1)
				}
				return nil, facade.ErrBadRequest.Facade(errMsgStr)
			}
		}
	}

	// 对于其他类型（图片、语音、缩略图），返回二进制数据
	mediaData, err := io.ReadAll(mediaResponse.Body)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.WechatMediaResponse{
		ContentType: contentType,
		Content:     mediaData,
	}, nil
}
