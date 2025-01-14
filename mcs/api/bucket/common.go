package bucket

import (
	"fmt"
	"go-mcs-sdk/mcs/api/common/constants"
	"go-mcs-sdk/mcs/api/common/utils"
	"go-mcs-sdk/mcs/api/common/web"
	"strings"

	"go-mcs-sdk/mcs/api/common/logs"
)

func (bucketClient *BucketClient) GetGateway() (*string, error) {
	apiUrl := utils.UrlJoin(bucketClient.BaseUrl, constants.API_URL_BUCKET_GATEWAY_GET_GATEWAY)

	var subDomains []string

	err := web.HttpGet(apiUrl, bucketClient.JwtToken, nil, &subDomains)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(subDomains) <= 0 {
		err := fmt.Errorf("no gateway returned")
		logs.GetLogger().Error(err)
		return nil, err
	}

	gateway := subDomains[0]
	if !strings.HasPrefix(gateway, "http") {
		gateway = "https://" + gateway
	}

	return &gateway, nil
}
