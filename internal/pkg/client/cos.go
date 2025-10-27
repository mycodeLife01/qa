package client

import (
	"net/http"
	"net/url"

	"github.com/mycodeLife01/qa/config"
	"github.com/tencentyun/cos-go-sdk-v5"
)

func InitCosClient() *cos.Client {
	u, _ := url.Parse("https://my-qa-go-1313494932.cos.ap-shanghai.myqcloud.com")
	su, _ := url.Parse("https://cos.ap-shanghai.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.C.COS.SecretID,
			SecretKey: config.C.COS.SecretKey,
		},
	})
	return client
}
