package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-mcs-sdk/mcs/common"
	"go-mcs-sdk/mcs/common/constants"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"unsafe"

	"github.com/ethereum/go-ethereum/crypto"
)

type McsClient struct {
	Client
	UserWalletAddressForRegisterMcs string `json:"user_wallet_address_for_register_mcs"`
	UserWalletAddressPK             string `json:"user_wallet_address_pk"`
	ChainNameForRegisterOnMcs       string `json:"chain_name_for_register_on_mcs"`
}

func NewMcsClient() *McsClient {
	mcsClient := McsClient{}
	mcsClient = *mcsClient.GetConfig()
	return &mcsClient
}

func (client *McsClient) GetConfig() *McsClient {
	err := common.LoadEnv()
	if err != nil {
		log.Fatal(err)
		return client
	}
	walletAddress := os.Getenv("USER_WALLET_ADDRESS_FOR_REGISTER_MCS")
	if walletAddress == "" {
		err = fmt.Errorf("user wallet address is null in .env file")
		log.Fatal(err)
		return client
	}
	client.UserWalletAddressForRegisterMcs = walletAddress
	walletAddressPK := os.Getenv("USER_WALLET_ADDRESS_PK")
	if walletAddressPK == "" {
		err = fmt.Errorf("user wallet address private key is null in .env file")
		log.Fatal(err)
		return client
	}
	client.UserWalletAddressPK = walletAddressPK
	chainNetworkName := os.Getenv("CHAIN_NAME_FOR_REGISTER_ON_MCS")
	if chainNetworkName == "" {
		err = fmt.Errorf("chain network name is null in .env file")
		log.Fatal(err)
		return client
	}
	client.ChainNameForRegisterOnMcs = chainNetworkName
	mcsBackendBaseUrl := os.Getenv("MCS_BACKEND_BASE_URL")
	if mcsBackendBaseUrl == "" {
		err = fmt.Errorf("mcs backend base url is null in .env file")
		log.Fatal(err)
		return client
	}
	client.BaseURL = mcsBackendBaseUrl
	return client
}

func (client *McsClient) UserLogin(walletAddress, signature, nonce, network string) ([]byte, error) {
	httpRequestUrl := client.BaseURL + constants.USER_LOGIN
	params := make(map[string]string)
	params["public_key_address"] = walletAddress
	params["nonce"] = nonce
	params["signature"] = signature
	params["network"] = network
	response, err := common.HttpPost(httpRequestUrl, client.JwtToken, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(*(*string)(unsafe.Pointer(&response)))
	return response, nil
}

func (client *McsClient) UserRegister(walletAddress string) (*string, error) {
	httpRequestUrl := client.BaseURL + constants.USER_REGISTER
	params := make(map[string]string)
	params["public_key_address"] = walletAddress
	response, err := common.HttpPost(httpRequestUrl, client.JwtToken, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(*(*string)(unsafe.Pointer(&response)))
	var dict map[string]interface{}
	err = json.Unmarshal(response, &dict)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(dict)
	dataInReturn := dict["data"].(map[string]interface{})
	objectInData := dataInReturn["nonce"].(string)
	fmt.Println(objectInData)
	return &objectInData, nil
}

func (client *McsClient) GetJwtToken() error {
	nonce, err := client.UserRegister(client.UserWalletAddressForRegisterMcs)
	if err != nil {
		log.Println(err)
		return err
	}
	privateKey, _ := crypto.HexToECDSA(client.UserWalletAddressPK)
	signature, _ := common.PersonalSign(*nonce, privateKey)
	resp, err := client.UserLogin(client.UserWalletAddressForRegisterMcs, signature, *nonce, client.ChainNameForRegisterOnMcs)
	if err != nil {
		log.Println(err)
		return err
	}
	var dict map[string]interface{}
	err = json.Unmarshal(resp, &dict)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(dict)
	dataInReturn := dict["data"].(map[string]interface{})
	jwtToken := dataInReturn["jwt_token"].(string)
	client.SetJwtToken(jwtToken)
	return nil
}

func (client *McsClient) GetUserTasksDeals(fileName, status string, pageNumber, pageSize int) ([]byte, error) {
	requestParam := "?file_name=" + fileName + "status=" + status + "page_number=" + strconv.Itoa(pageNumber) + "page_size=" + strconv.Itoa(pageSize)
	httpRequestUrl := client.BaseURL + constants.TASKS_DEALS + requestParam
	response, err := common.HttpGet(httpRequestUrl, client.JwtToken, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(*(*string)(unsafe.Pointer(&response)))
	return response, nil
}

func (client *McsClient) GetDealDetail(sourceFileUploadId, dealId int) ([]byte, error) {
	requestParam := strconv.Itoa(dealId) + "?source_file_upload_id=" + strconv.Itoa(sourceFileUploadId)
	httpRequestUrl := client.BaseURL + constants.DEAL_DETAIL + requestParam
	response, err := common.HttpGet(httpRequestUrl, client.JwtToken, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(*(*string)(unsafe.Pointer(&response)))
	return response, nil
}

func (client *McsClient) GetMintInfo(sourceFileUploadId, tokenId int, payloadCid, txHash, mintAddress string) ([]byte, error) {
	httpRequestUrl := client.BaseURL + constants.MINT_INFO
	params := make(map[string]interface{})
	params["source_file_upload_id"] = sourceFileUploadId
	params["payload_cid"] = payloadCid
	params["tx_hash"] = txHash
	params["token_id"] = tokenId
	params["mint_address"] = mintAddress
	response, err := common.HttpPost(httpRequestUrl, client.JwtToken, params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(*(*string)(unsafe.Pointer(&response)))
	return response, nil
}

func (client *McsClient) UploadFile(filePath string) ([]byte, error) {
	httpRequestUrl := client.BaseURL + constants.UPLOAD_FILE
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(filePath)
	defer file.Close()
	part1, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if errFile1 != nil {
		fmt.Println(errFile1)
		return nil, err
	}
	_, err = io.Copy(part1, file)
	if err != nil {
		fmt.Println(errFile1)
		return nil, err
	}
	err = writer.WriteField("duration", "525")
	if err != nil {
		fmt.Println(errFile1)
		return nil, err
	}
	err = writer.WriteField("storage_copy", "5")
	if err != nil {
		fmt.Println(errFile1)
		return nil, err
	}
	err = writer.WriteField("wallet_address", client.UserWalletAddressForRegisterMcs)
	if err != nil {
		fmt.Println(errFile1)
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", httpRequestUrl, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.JwtToken))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(string(body))
	return body, nil
}

func (client *McsClient) GenerateApikey(validDays int) ([]byte, error) {
	httpRequestUrl := client.BaseURL + constants.GENERATE_APIKEY + strconv.Itoa(validDays)
	response, err := common.HttpGet(httpRequestUrl, client.JwtToken, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(*(*string)(unsafe.Pointer(&response)))
	return response, nil
}