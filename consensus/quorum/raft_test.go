package quorum

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ConsenSys/quorum-hibernate/config"
	"github.com/stretchr/testify/require"
)

func TestRaftConsensus_ValidateShutdown_Minter_Invalid(t *testing.T) {
	mockServer := startMockRaftServer(t, `{"result": "minter"}`, "")
	defer mockServer.Close()

	raft := NewRaftConsensus(&config.Node{
		BasicConfig: &config.Basic{
			BlockchainClient: &config.BlockchainClient{
				BcClntRpcUrl: mockServer.URL,
			},
		},
	}, nil)

	isConsensusNode, err := raft.ValidateShutdown()
	require.EqualError(t, err, "minter node, cannot be shutdown")
	require.True(t, isConsensusNode)
}

func TestRaftConsensus_ValidateShutdown_Learner_Valid(t *testing.T) {
	mockServer := startMockRaftServer(t, `{"result": "learner"}`, "")
	defer mockServer.Close()

	raft := NewRaftConsensus(&config.Node{
		BasicConfig: &config.Basic{
			BlockchainClient: &config.BlockchainClient{
				BcClntRpcUrl: mockServer.URL,
			},
		},
	}, nil)

	isConsensusNode, err := raft.ValidateShutdown()
	require.NoError(t, err)
	require.False(t, isConsensusNode)
}

func TestRaftConsensus_ValidateShutdown_Verifier_NotEnoughActivePeers_Invalid(t *testing.T) {
	var tests = []struct {
		name, raftRoleResp, raftClusterResp string
		wantErrMsg                          string
	}{
		{
			name:            "notEnoughPeers",
			raftRoleResp:    `{"result": "verifier"}`,
			raftClusterResp: `{"result": [{"NodeActive":true},{"NodeActive":true}]}`,
			wantErrMsg:      "raft quorum failed, activeNodes=2 minimumActiveNodesRequired=2 cannot be shutdown",
		},
		{
			name:            "notEnoughActivePeers",
			raftRoleResp:    `{"result": "verifier"}`,
			raftClusterResp: `{"result": [{"NodeActive":true},{"NodeActive":true},{"NodeActive":false}]}`,
			wantErrMsg:      "raft quorum failed, activeNodes=2 minimumActiveNodesRequired=2 cannot be shutdown",
		},
		{
			name:            "enoughActivePeers",
			raftRoleResp:    `{"result": "verifier"}`,
			raftClusterResp: `{"result": [{"NodeActive":true},{"NodeActive":true},{"NodeActive":true}]}`,
			wantErrMsg:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := startMockRaftServer(t, tt.raftRoleResp, tt.raftClusterResp)
			defer mockServer.Close()

			raft := NewRaftConsensus(&config.Node{
				BasicConfig: &config.Basic{
					BlockchainClient: &config.BlockchainClient{
						BcClntRpcUrl: mockServer.URL,
					},
				},
			}, nil)

			isConsensusNode, err := raft.ValidateShutdown()
			if tt.wantErrMsg == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErrMsg)
			}
			require.True(t, isConsensusNode)
		})
	}
}

func TestRaftConsensus_ValidateShutdown_GetRoleRpcError(t *testing.T) {
	var (
		raftRoleResp    = `{"error": {"code":111,"message":"someerror","data":{"additional":"context"}}}`
		raftClusterResp = `{"result": [{"NodeActive":true},{"NodeActive":true},{"NodeActive":true}]}`
		wantErrMsg      = "unable to check raft role: code = 111, message = someerror, data = map[additional:context]"
	)

	mockServer := startMockRaftServer(t, raftRoleResp, raftClusterResp)
	defer mockServer.Close()

	raft := NewRaftConsensus(&config.Node{
		BasicConfig: &config.Basic{
			BlockchainClient: &config.BlockchainClient{
				BcClntRpcUrl: mockServer.URL,
			},
		},
	}, nil)

	isConsensusNode, err := raft.ValidateShutdown()

	require.EqualError(t, err, wantErrMsg)
	require.False(t, isConsensusNode)
}

func TestRaftConsensus_ValidateShutdown_GetClusterInfoRpcError(t *testing.T) {
	var (
		raftRoleResp    = `{"result": "verifier"}`
		raftClusterResp = `{"error": {"code":111,"message":"someerror","data":{"additional":"context"}}}`
		wantErrMsg      = "unable to check raft cluster info: code = 111, message = someerror, data = map[additional:context]"
	)

	mockServer := startMockRaftServer(t, raftRoleResp, raftClusterResp)
	defer mockServer.Close()

	raft := NewRaftConsensus(&config.Node{
		BasicConfig: &config.Basic{
			BlockchainClient: &config.BlockchainClient{
				BcClntRpcUrl: mockServer.URL,
			},
		},
	}, nil)

	isConsensusNode, err := raft.ValidateShutdown()

	require.EqualError(t, err, wantErrMsg)
	require.True(t, isConsensusNode)
}

func startMockRaftServer(t *testing.T, raftRoleResp, raftClusterResp string) *httptest.Server {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		type rpcRequest struct {
			Method string
		}

		rpcReq := rpcRequest{}

		err := json.NewDecoder(req.Body).Decode(&rpcReq)
		require.NoError(t, err)

		if rpcReq.Method == "raft_role" {
			_, err := io.WriteString(w, raftRoleResp)
			require.NoError(t, err)
		} else if rpcReq.Method == "raft_cluster" {
			_, err := io.WriteString(w, raftClusterResp)
			require.NoError(t, err)
		}
	})

	return httptest.NewServer(serverMux)
}
