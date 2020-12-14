package config

import (
	"errors"
	"fmt"
)

type RPCServer struct {
	RpcAddr     string     `toml:"rpcAddr"`
	RPCCorsList []string   `toml:"rpcCorsList"`
	RPCVHosts   []string   `toml:"rpcvHosts"`
	TLSConfig   *ServerTLS `toml:"tlsConfig"`
}

func (c RPCServer) IsValid() error {
	if c.RpcAddr == "" {
		return errors.New("rpcAddr is empty")
	}
	if err := isValidUrl(c.RpcAddr); err != nil {
		return fmt.Errorf("invalid rpcAddr: %v", err)
	}
	if c.TLSConfig != nil {
		if err := c.TLSConfig.IsValid(); err != nil {
			return err
		}
	}
	return nil
}
