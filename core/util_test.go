package core

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetRandomRetryWaitTime(t *testing.T) {
	c := 1
	for c <= 1000 {
		w := GetRandomRetryWaitTime(100, 1000)
		if w > 1000 || w < 100 {
			t.Error("wait time is out of range (100 - 1000)")
		}
		c++
	}
}

func TestCallRPC(t *testing.T) {
	var (
		rpcMethod = "app.DoSomething"
		req       = fmt.Sprintf(`{"jsonrpc":2.0, "id":11, "method":"%v"}`, rpcMethod)
		respCode  = 200
		respBody  = `{"jsonrpc":"2.0","id":1,"result":{"someresponsedata": "val"}}`
	)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, req.Method, "POST")
		require.Equal(t, req.Header["Content-Type"], []string{"application/json"})

		type rpcRequest struct {
			Method string
		}

		rpcReq := rpcRequest{}

		err := json.NewDecoder(req.Body).Decode(&rpcReq)
		require.NoError(t, err)
		require.Equal(t, rpcMethod, rpcReq.Method)

		w.WriteHeader(respCode)
		_, err = w.Write([]byte(respBody))
		require.NoError(t, err)
	})

	mockServer := httptest.NewServer(serverMux)

	var got interface{}

	err := CallRPC(mockServer.URL, []byte(req), &got)
	require.NoError(t, err)

	rpcResp := got.(map[string]interface{})
	rpcResult := rpcResp["result"].(map[string]interface{})

	require.Contains(t, rpcResult, "someresponsedata")
}

func TestCallRPC_HTTPError(t *testing.T) {
	var (
		rpcMethod = "app.DoSomething"
		req       = fmt.Sprintf(`{"jsonrpc":2.0, "id":11, "method":"%v"}`, rpcMethod)
	)

	var tests = []struct {
		name     string
		respCode int
		respBody string
		wantErr  string
	}{
		{name: "clientError", respCode: 400, wantErr: "CallRPC - response status failed, not OK, status=400 Bad Request"},
		{name: "serverError", respCode: 500, wantErr: "CallRPC - response status failed, not OK, status=500 Internal Server Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverMux := http.NewServeMux()
			serverMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

				require.Equal(t, req.Method, "POST")
				require.Equal(t, req.Header["Content-Type"], []string{"application/json"})

				type rpcRequest struct {
					Method string
				}

				rpcReq := rpcRequest{}

				err := json.NewDecoder(req.Body).Decode(&rpcReq)
				require.NoError(t, err)
				require.Equal(t, rpcMethod, rpcReq.Method)

				w.WriteHeader(tt.respCode)
				_, err = w.Write([]byte(tt.respBody))
				require.NoError(t, err)
			})

			mockServer := httptest.NewServer(serverMux)

			var resp interface{}

			err := CallRPC(mockServer.URL, []byte(req), &resp)
			require.EqualError(t, err, tt.wantErr)
			require.Empty(t, resp)
		})
	}
}

func TestCallRPC_RpcError(t *testing.T) {
	var (
		rpcMethod = "app.DoSomething"
		req       = fmt.Sprintf(`{"jsonrpc":2.0, "id":11, "method":"%v"}`, rpcMethod)
		respCode  = 200
		respBody  = `{"jsonrpc":"2.0","id":1,"error":{"code":100,"message":"someerrormessage", "data":{"field": "val"}}}`
	)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		require.Equal(t, req.Method, "POST")
		require.Equal(t, req.Header["Content-Type"], []string{"application/json"})

		type rpcRequest struct {
			Method string
		}

		rpcReq := rpcRequest{}

		err := json.NewDecoder(req.Body).Decode(&rpcReq)
		require.NoError(t, err)
		require.Equal(t, rpcMethod, rpcReq.Method)

		w.WriteHeader(respCode)
		_, err = w.Write([]byte(respBody))
		require.NoError(t, err)
	})

	mockServer := httptest.NewServer(serverMux)

	var got interface{}

	err := CallRPC(mockServer.URL, []byte(req), &got)
	require.NoError(t, err)

	rpcResp := got.(map[string]interface{})
	rpcErr := rpcResp["error"].(map[string]interface{})
	require.Equal(t, rpcErr["message"], "someerrormessage")

	rpcErrData := rpcErr["data"].(map[string]interface{})
	require.Contains(t, rpcErrData, "field")
}

func TestCallRPC_InvalidRespType(t *testing.T) {
	var (
		rpcMethod = "app.DoSomething"
		req       = fmt.Sprintf(`{"jsonrpc":2.0, "id":11, "method":"%v"}`, rpcMethod)
		respCode  = 200
		respBody  = `{"jsonrpc":"2.0","id":1,"result":{"someresponsedata": "val"}}`
	)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		require.Equal(t, req.Method, "POST")
		require.Equal(t, req.Header["Content-Type"], []string{"application/json"})

		type rpcRequest struct {
			Method string
		}

		rpcReq := rpcRequest{}

		err := json.NewDecoder(req.Body).Decode(&rpcReq)
		require.NoError(t, err)
		require.Equal(t, rpcMethod, rpcReq.Method)

		w.WriteHeader(respCode)
		_, err = w.Write([]byte(respBody))
		require.NoError(t, err)
	})

	mockServer := httptest.NewServer(serverMux)

	type invalid struct {
		NotKnown int
	}

	var got invalid

	err := CallRPC(mockServer.URL, []byte(req), &got)
	require.NoError(t, err)
	require.Empty(t, got)
}
