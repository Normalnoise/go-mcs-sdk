package api

import (
	"go-mcs-sdk/mcs/api/user"
	"go-mcs-sdk/mcs/config"
	"testing"

	"go-mcs-sdk/mcs/api/common/logs"

	"github.com/stretchr/testify/assert"
)

var onChainClient *OnChainClient

func init() {
	if onChainClient != nil {
		return
	}

	apikey := config.GetConfig().Apikey
	accessToken := config.GetConfig().AccessToken
	network := config.GetConfig().Network

	mcsClient, err := user.LoginByApikey(apikey, accessToken, network)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	onChainClient = GetOnChainClient(*mcsClient)
}

func TestGetSystemParam(t *testing.T) {
	params, err := onChainClient.GetSystemParam()
	assert.Nil(t, err)
	assert.NotEmpty(t, params)

	logs.GetLogger().Info(params)
}

func TestGetFilPrice(t *testing.T) {
	price, err := GetHistoricalAveragePriceVerified()
	assert.Nil(t, err)
	assert.NotNil(t, price)
	assert.GreaterOrEqual(t, price, float64(0))

	logs.GetLogger().Info(price)
}

func TestGetAmount(t *testing.T) {
	amount, err := GetAmount(1, 0.1, 1, 2)
	assert.Nil(t, err)
	assert.NotEmpty(t, amount)
	assert.GreaterOrEqual(t, amount, int64(0))

	logs.GetLogger().Info(amount)
}