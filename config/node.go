package config

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ConsenSys/quorum-hibernate/core"
	"github.com/ConsenSys/quorum-hibernate/log"
)

type Node struct {
	BasicConfig *Basic  `toml:"basicConfig" json:"basicConfig"` // basic config of this node hibernator
	Peers       PeerArr // node hibernator config of other node hibernators
}

func (c Node) IsConsensusValid(client *http.Client) error {
	const (
		adminInfoReq = `{"jsonrpc":"2.0", "method":"admin_nodeInfo", "params":[], "id":67}`
		protocolKey  = "protocols"
		ethKey       = "eth"
		consensusKey = "consensus"
		istanbulKey  = "istanbul"
	)
	log.Debug("IsConsensusValid - validating consensus info")

	if c.BasicConfig.IsBesuClient() {
		return nil
	}

	var resp map[string]interface{}
	if err := core.CallRPC(client, c.BasicConfig.BlockchainClient.BcClntRpcUrl, []byte(adminInfoReq), &resp); err == nil {
		resMap := resp["result"].(map[string]interface{})
		log.Debug("IsConsensusValid - response", "map", resMap)

		if resMap[protocolKey] == nil {
			return errors.New("IsConsensusValid - no consensus info found")
		}
		protocols, ok := resMap[protocolKey].(map[string]interface{})
		if !ok {
			return errors.New("IsConsensusValid - invalid consensus info found")
		}
		if protocols[istanbulKey] != nil {
			if c.BasicConfig.IsIstanbul() {
				return nil
			}
			return errors.New("IsConsensusValid - invalid consensus. it should be istanbul")
		}
		eth := protocols[ethKey].(map[string]interface{})
		if _, ok := eth[consensusKey]; !ok {
			return fmt.Errorf("IsConsensusValid - consensus key missing in node info api output")
		} else {
			expected := eth[consensusKey].(string)
			log.Debug("IsConsensusValid - consensus name", "name", expected)
			if expected == c.BasicConfig.BlockchainClient.Consensus {
				return nil
			}
			return fmt.Errorf("IsConsensusValid - consensus mismatch. expected:%s, have:%s", expected, c.BasicConfig.BlockchainClient.Consensus)
		}
	}
	return nil
}
