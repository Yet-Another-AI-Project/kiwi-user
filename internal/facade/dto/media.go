package dto

// WechatMediaResponse represents the response for WeChat media resource
type WechatMediaResponse struct {
	// For video type
	VideoURL string `json:"video_url,omitempty"` // 视频下载地址，ContentType为video时有效

	// For image/voice/thumb type
	ContentType string `json:"content_type,omitempty"` // 媒体类型，如audio/mp3、image/jpeg、image/png、image/gif、image/jpg
	Content     []byte `json:"content,omitempty"`      // 实际的媒体内容
}
