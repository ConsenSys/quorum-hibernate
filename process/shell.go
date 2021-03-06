package process

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ConsenSys/quorum-hibernate/config"

	"github.com/ConsenSys/quorum-hibernate/log"
)

// ShellProcessControl represents process control for a shell process
type ShellProcessControl struct {
	cfg     *config.Process
	status  bool
	client  *http.Client
	muxLock sync.Mutex
}

func NewShellProcess(c *http.Client, p *config.Process, s bool) Process {
	sp := &ShellProcessControl{p, s, c, sync.Mutex{}}
	sp.UpdateStatus()
	log.Debug("shell process created", "name", sp.cfg.Name)
	return sp
}

func (sp *ShellProcessControl) setStatus(s bool) {
	sp.status = s
	log.Debug("setStatus - process "+sp.cfg.Name, "status", sp.status)
}

// Status implements Process.Status
func (sp *ShellProcessControl) Status() bool {
	return sp.status
}

// UpdateStatus implements Process.UpdateStatus
func (sp *ShellProcessControl) UpdateStatus() bool {

	s := false
	var err error
	s, err = IsProcessUp(sp.client, sp.cfg.UpcheckCfg)
	if err != nil {
		sp.setStatus(false)
		log.Error("UpdateStatus - shell process is down", "err", err)
	} else {
		sp.setStatus(s)
	}
	log.Debug("UpdateStatus", "name", sp.cfg.Name, "return", sp.status)
	return sp.status
}

// Status implements Process.Stop
func (sp *ShellProcessControl) Stop() error {
	defer sp.muxLock.Unlock()
	sp.muxLock.Lock()
	if !sp.status {
		log.Debug("Stop - process is already down", "name", sp.cfg.Name)
		return nil
	}
	if err := ExecuteShellCommand(sp.cfg.StopCommand); err == nil {
		if sp.WaitToBeDown() {
			sp.setStatus(false)
			log.Debug("Stop - stopped", "process", sp.cfg.Name, "status", sp.status)
		} else {
			sp.setStatus(true)
			log.Error("Stop - failed to stop " + sp.cfg.Name)
			return fmt.Errorf("Stop - %s failed to stop", sp.cfg.Name)
		}
		log.Debug("Stop - stopped", "process", sp.cfg.Name, "status", sp.status)
	} else {
		log.Error("Stop - "+sp.cfg.Name+" failed", "err", err)
		return err
	}
	return nil
}

// Status implements Process.Start
func (sp *ShellProcessControl) Start() error {
	defer sp.muxLock.Unlock()
	sp.muxLock.Lock()
	if sp.status {
		log.Info("Start - process is already up", "name", sp.cfg.Name)
		return nil
	}
	if err := ExecuteShellCommand(sp.cfg.StartCommand); err == nil {
		//wait for process to come up
		if sp.WaitToComeUp() {
			sp.setStatus(true)
			log.Debug("Start - started", "process", sp.cfg.Name, "status", sp.status)
		} else {
			sp.setStatus(false)
			log.Error("Start - failed to start " + sp.cfg.Name)
			return fmt.Errorf("%s failed to start", sp.cfg.Name)
		}

	} else {
		log.Error("Start - failed to start " + sp.cfg.Name)
		return err
	}
	return nil
}

// TODO create helper method that can be called from  docker as well
// WaitToComeUp waits for the process status to be up by performing up check repeatedly
// for a certain duration
func (sp *ShellProcessControl) WaitToComeUp() bool {
	retryCount := 30
	c := 1
	for c <= retryCount {
		if sp.UpdateStatus() {
			return true
		}
		time.Sleep(time.Second)
		log.Debug("WaitToComeUp - wait for up "+sp.cfg.Name, "c", c)
		c++
	}
	return false
}

// TODO create helper that can be called from  docker as well
// WaitToBeDown waits for the process status to be down by performing up check repeatedly
// for a certain duration
func (sp *ShellProcessControl) WaitToBeDown() bool {
	retryCount := 30
	c := 1
	for c <= retryCount {
		if !sp.UpdateStatus() {
			return true
		}
		time.Sleep(time.Second)
		log.Info("WaitToBeDown - wait for down "+sp.cfg.Name, "c", c)
		c++
	}
	return false
}
