package api

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/ElrondNetwork/elrond-go/logger"
)

type sendTxRequest struct {
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Value     string `json:"value"`
	Data      string `json:"data"`
	Nonce     uint64 `json:"nonce"`
	GasPrice  uint64 `json:"gasPrice"`
	GasLimit  uint64 `json:"gasLimit"`
	Signature string `json:"signature"`
}

type sendTxResponse struct {
	TxHash string `json:"txHash"`
	Error  string `json:"Error"`
}

type account struct {
	Address string `json:"address"`
	Nonce   uint64 `json:"nonce"`
	Balance string `json:"balance"`
	//Code     string
	//CodeHash []byte
	//RootHash []byte
}

type accountWrapper struct {
	Account account `json:"account"`
}

var (
	urlValidationRegexp = regexp.MustCompile(`^http(s)?:\/\/wallet\-api\.elrond\.com`)
)

// SendTransaction performs the actual HTTP request to send the transaction
func SendTransaction(
	apiHost string,
	nonce uint64,
	sender string,
	receiver string,
	value string,
	gasPrice uint64,
	gasLimit uint64,
	data string,
	signature []byte,
	proxy string,
	log logger.Logger) (string, error) {

	//url := "https://wallet-api.elrond.com/transaction/send"
	url := fmt.Sprintf("%s/transaction/send", apiHost)
	hexSignature := hex.EncodeToString(signature)
	log.Info(fmt.Sprintf("Signature for tx is: %s", hexSignature))

	txReq := sendTxRequest{
		Sender:    sender,
		Receiver:  receiver,
		Value:     value,
		Data:      data,
		Nonce:     nonce,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Signature: hexSignature,
	}

	jsonData, _ := json.Marshal(txReq)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	log.Info(fmt.Sprintf("Sending tx request to %s", url))

	body, err := performRequest(url, proxy, req, log)

	if err != nil {
		return "", err
	}

	var response sendTxResponse
	json.Unmarshal([]byte(body), &response)

	if response.TxHash == "" {
		return "", fmt.Errorf("transaction failed: %s", response)
	}

	return response.TxHash, nil
}

// GetAccount fetches the desired account's balance as well as nonce
func GetAccount(apiHost string, address string, proxy string, log logger.Logger) (account, error) {
	url := fmt.Sprintf("https://wallet-api.elrond.com/address/%s", address)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var response accountWrapper
	var accountResponse account

	body, _ := performRequest(url, proxy, req, log)

	if err != nil {
		return accountResponse, err
	}

	json.Unmarshal([]byte(body), &response)
	accountResponse = response.Account

	return accountResponse, nil
}

func performRequest(requestURL string, proxy string, request *http.Request, log logger.Logger) ([]byte, error) {
	isElrondAPI := isElrondWalletAPI(requestURL)

	client := &http.Client{}

	if isElrondAPI && proxy != "" {
		log.Info(fmt.Sprintf("Will use proxy %s to request %s", proxy, requestURL))

		proxyURL, _ := url.Parse(proxy)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyURL),
		}

		client.Transport = transport
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Error(fmt.Sprintf("Request to url %s failed!", requestURL), err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Request to url %s failed!", requestURL), err)
		return nil, err
	}

	defer resp.Body.Close()

	return body, err
}

func isElrondWalletAPI(url string) bool {
	matches := urlValidationRegexp.FindAllStringSubmatch(url, -1)
	if len(matches) > 0 {
		return true
	}

	return false
}
