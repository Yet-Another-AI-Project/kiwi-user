package config

type OSSConfig struct {
	AccessKeyID     string `config:"access_key_id"`
	AccessKeySecret string `config:"access_key_secret"`
	Endpoint        string `config:"endpoint"`
	BucketName      string `config:"bucket_name"`
	CDN             string `config:"cdn"`
}
