package main

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

var MasterConfigErr error
var MasterConf = new(MasterConfig)
var confInitOnce sync.Once

type MasterConfig struct {
	LogPath       string   `json:logPath`
	WebPort       string   `json:webServer`
	ReadTimeOut   int      `json:readTimeOut`
	WriteTimeOut  int      `json:writeTimeOut`
	EtcdEndPoints []string `json:etcdEndPoints`
	MongoDB       string   `json:mongoDB`
}

func MasterConfigInit() {
	confInitOnce.Do(masterConfigInitOnce)
}

func masterConfigInitOnce() {
	var (
		confContent []byte
		err         error
		confPath    = ArgumentConfPath
	)

	if confContent, err = ioutil.ReadFile(confPath); err != nil {
		MasterConfigErr = err
		return
	}

	if err = json.Unmarshal(confContent, MasterConf); err != nil {
		MasterConfigErr = err
		return
	}
}
