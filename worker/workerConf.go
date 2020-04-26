package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

var WorkerConfigErr error
var WorkerConf = new(WorkerConfig)
var confInitOnce sync.Once

type WorkerConfig struct {
	LogPath       string   `json:logPath`
	WebPort       string   `json:webServer`
	ReadTimeOut   int      `json:readTimeOut`
	WriteTimeOut  int      `json:writeTimeOut`
	EtcdEndPoints []string `json:etcdEndPoints`
	MongoDB       string   `json:mongoDB`
}

func (w *WorkerConfig) WorkerConfigInit() {
	confInitOnce.Do(workerConfigInitOnce)
}

func workerConfigInitOnce() {
	var (
		confContent []byte
		err         error
		confPath    = ConfPath
	)

	if confContent, err = ioutil.ReadFile(confPath); err != nil {
		WorkerConfigErr = err
		return
	}

	if err = json.Unmarshal(confContent, WorkerConf); err != nil {
		WorkerConfigErr = err
		return
	}
}
