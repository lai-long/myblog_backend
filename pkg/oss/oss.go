// Package oss 封装对 S3 兼容对象存储（RustFS 等）的访问。
// 只负责“连上”和“上传/拼 URL”，业务校验放在调用方。
package oss

import (
	"bytes"
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config 与 blog 的 OSS 配置一一对应（独立定义，避免 pkg 依赖 internal）。
type Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
	PublicBaseURL   string
}

// Client 持有 S3 客户端和桶信息。
type Client struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

// New 根据配置创建 Client。自建 S3 兼容服务一般要 UsePathStyle=true 并指定 Endpoint。
func New(cfg Config) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &Client{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

// Upload 把二进制内容上传到 key 指定的对象，返回可公开访问的 URL。
func (c *Client) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return c.PublicURL(key), nil
}

// PublicURL 由 key 拼出对象的对外访问地址。
func (c *Client) PublicURL(key string) string {
	return c.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}
