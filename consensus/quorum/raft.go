package quorum

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ConsenSysQuorum/node-manager/consensus"
	"github.com/ConsenSysQuorum/node-manager/core"
	"github.com/ConsenSysQuorum/node-manager/core/types"
	"github.com/ConsenSysQuorum/node-manager/log"
)

// RaftClusterEntry represents entries from the output of rpc method raft_cluster
type RaftClusterEntry struct {
	Hostname   string `json:"hostName"`
	NodeActive bool   `json:"nodeActive"`
	NodeId     string `json:"nodeId"`
	P2pPort    int    `json:"p2pPort"`
	RaftId     int    `json:"raftId"`
	Role       string `json:"role"`
}

type RaftClusterResp struct {
	Result []RaftClusterEntry `json:"result"`
	Error  error              `json:"error"`
}

type RaftRoleResp struct {
	Result string `json:"result"`
	Error  error  `json:"error"`
}

type RaftConsensus struct {
	cfg    *types.NodeConfig
	client *http.Client
}

const (
	// Raft roles
	MINTER  = "minter"
	LEARNER = "learner"

	// Raft RPC apis
	RaftRoleReq    = `{"jsonrpc":"2.0", "method":"raft_role", "params":[], "id":67}`
	RaftClusterReq = `{"jsonrpc":"2.0", "method":"raft_cluster", "params":[], "id":67}`
)

func NewRaftConsensus(qn *types.NodeConfig) consensus.Consensus {
	return &RaftConsensus{cfg: qn, client: core.NewHttpClient()}
}

func (r *RaftConsensus) getRole(rpcUrl string) (string, error) {
	var respResult RaftRoleResp
	if err := core.CallRPC(rpcUrl, []byte(RaftRoleReq), &respResult); err != nil {
		return "", err
	}
	return respResult.Result, respResult.Error
}

func (r *RaftConsensus) getRaftClusterInfo(rpcUrl string) ([]RaftClusterEntry, error) {
	var respResult RaftClusterResp
	if err := core.CallRPC(rpcUrl, []byte(RaftClusterReq), &respResult); err != nil {
		return nil, err
	}
	return respResult.Result, respResult.Error
}

// ValidateShutdown implements Consensus.ValidateShutdown
func (r *RaftConsensus) ValidateShutdown() error {
	role, err := r.getRole(r.cfg.BasicConfig.BcClntRpcUrl)
	if err != nil {
		log.Error("ValidateShutdown - raft role failed", "err", err)
		return err
	}

	if role == MINTER {
		return errors.New("ValidateShutdown - minter node, cannot be shutdown")
	}

	if role == LEARNER {
		log.Info("ValidateShutdown - raft consensus check - role:learner, ok to shutdown")
		return nil
	}

	cluster, err := r.getRaftClusterInfo(r.cfg.BasicConfig.BcClntRpcUrl)
	if err != nil {
		log.Error("ValidateShutdown - raft cluster failed", "err", err)
		return err
	}

	activeNodes := 0
	totalNodes := len(cluster)
	for _, n := range cluster {
		if n.NodeActive {
			activeNodes++
		}
	}
	minActiveNodes := (totalNodes / 2) + 1
	log.Info("ValidateShutdown - raft consensus check", "role", role, "minActiveNodes", minActiveNodes, "totalNodes", totalNodes, "ActiveNodes", activeNodes)

	if activeNodes <= minActiveNodes {
		return fmt.Errorf("ValidateShutdown - raft quorum failed, activeNodes=%d minimmumActiveNodesRequired=%d cannot be shutdown", activeNodes, minActiveNodes)
	}
	return nil
}