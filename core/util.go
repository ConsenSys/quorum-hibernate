package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/ConsenSys/quorum-hibernate/log"
)

var defaultClient *http.Client

func init() {
	defaultClient = NewHttpClient(nil)
}

// NewHttpClient returns a new customized http client
func NewHttpClient(tls *tls.Config) *http.Client {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: HttpClientRequestDialerTimeout,
		}).DialContext,
		TLSHandshakeTimeout: TLSHandshakeTimeout,
		TLSClientConfig:     tls,
	}
	var netClient = &http.Client{
		Timeout:   HttpClientRequestTimeout,
		Transport: netTransport,
	}
	return netClient
}

// RandomInt returns a random int within a range of min to max
func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func CallRPC(client *http.Client, rpcUrl string, rpcReq []byte, resData interface{}) error {
	_, err := httpRequest(client, rpcUrl, "POST", rpcReq, resData, false)
	return err
}

func CallREST(client *http.Client, rpcUrl string, method string, rpcReq []byte) (string, error) {
	return httpRequest(client, rpcUrl, method, rpcReq, nil, true)
}

// httpRequest makes a http request to rpcUrl. It makes http req with rpcReq as body.
// The returned JSON result is decoded into resData.
// resData must be a pointer.
// If http request returns 200 OK, it returns response body decoded into resData
// It returns error if http request does not return 200 OK or json decoding of response fails
// if returnRaw is true it returns the response as string and does not set resData
func httpRequest(client *http.Client, rpcUrl string, method string, rpcReq []byte, resData interface{}, returnRaw bool) (string, error) {
	if client == nil {
		client = defaultClient
	}
	log.Debug("CallRPC - making rpc call", "req", string(rpcReq))
	req, err := http.NewRequest(method, rpcUrl, bytes.NewBuffer(rpcReq))
	if err != nil {
		return "", fmt.Errorf("CallRPC - creating request failed err=%v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("CallRPC - do req failed err=%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("CallRPC - response status failed, not OK", "status", resp.Status)
		return "", fmt.Errorf("CallRPC - response status failed, not OK, status=%s", resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("CallRPC - response", "body", string(body))
	if returnRaw {
		return string(body), nil
	}
	if err := json.Unmarshal(body, resData); err != nil {
		log.Error("CallRPC - response json decode failed", "err", err)
		return "", err
	}

	log.Debug("CallRPC - response OK", "from", rpcUrl, "result", resData)
	return "", nil
}

type RpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (e *RpcError) Error() string {
	return fmt.Sprintf("code = %v, message = %v, data = %v", e.Code, e.Message, e.Data)
}
