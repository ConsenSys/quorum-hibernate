package config

import (
	"errors"
	"strings"
)

type BlockchainClient struct {
	ClientType      string     `toml:"type" json:"type"`           // client used by this node hibernator. it should be goquorum or besu
	Consensus       string     `toml:"consensus" json:"consensus"` // consensus used by blockchain client. ex: raft / istanbul / clique
	BcClntRpcUrl    string     `toml:"rpcUrl" json:"rpcUrl"`       // RPC url of blockchain client managed by this node hibernator
	BcClntTLSConfig *ClientTLS `toml:"tlsConfig" json:"tlsConfig"` // blockchain client TLS config
	BcClntProcess   *Process   `toml:"process" json:"process"`     // blockchain client process managed by this node hibernator
}

type PrivacyManager struct {
	PrivManKey       string     `toml:"publicKey" json:"publicKey"` // public key of privacy hibernator managed by this node hibernator
	PrivManTLSConfig *ClientTLS `toml:"tlsConfig" json:"tlsConfig"` // Privacy hibernator TLS config
	PrivManProcess   *Process   `toml:"process" json:"process"`     // privacy hibernator process managed by this node hibernator
}

func (c *BlockchainClient) IsRaft() bool {
	return strings.ToLower(c.Consensus) == "raft"
}

func (c *BlockchainClient) IsIstanbul() bool {
	return strings.ToLower(c.Consensus) == "istanbul"
}

func (c *BlockchainClient) IsClique() bool {
	return strings.ToLower(c.Consensus) == "clique"
}

func (c *BlockchainClient) IsGoQuorumClient() bool {
	return strings.ToLower(c.ClientType) == "goquorum"
}

func (c *BlockchainClient) IsBesuClient() bool {
	return strings.ToLower(c.ClientType) == "besu"
}

func (c *BlockchainClient) IsValid() error {
	if c.Consensus == "" {
		return newFieldErr("consensus", isEmptyErr)
	}

	if c.ClientType == "" {
		return newFieldErr("type", isEmptyErr)
	}
	if !c.IsGoQuorumClient() && !c.IsBesuClient() {
		return newFieldErr("type", errors.New("must be goquorum or besu"))
	}

	if c.IsGoQuorumClient() && !c.IsRaft() && !c.IsClique() && !c.IsIstanbul() {
		return newFieldErr("consensus", errors.New("must be raft, istanbul, or clique"))
	}

	if c.IsBesuClient() && !c.IsClique() {
		return newFieldErr("consensus", errors.New("must be clique"))
	}

	if c.BcClntRpcUrl == "" {
		return newFieldErr("rpcUrl", isEmptyErr)
	}

	if c.BcClntProcess == nil {
		return newFieldErr("process", isEmptyErr)
	}

	if err := c.BcClntProcess.IsValid(); err != nil {
		return newFieldErr("process", err)
	}

	if c.BcClntTLSConfig != nil {
		if err := c.BcClntTLSConfig.IsValid(); err != nil {
			return newFieldErr("tlsConfig", err)
		}
	}

	return nil
}

func (c *PrivacyManager) IsValid() error {
	if c.PrivManKey == "" {
		return newFieldErr("publicKey", isEmptyErr)
	}
	if c.PrivManProcess == nil {
		return newFieldErr("process", isEmptyErr)
	}
	if err := c.PrivManProcess.IsValid(); err != nil {
		return newFieldErr("process", err)
	}
	if c.PrivManTLSConfig != nil {
		if err := c.PrivManTLSConfig.IsValid(); err != nil {
			return newFieldErr("tlsConfig", err)
		}
	}

	return nil
}
