package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type EtcdEndpoint struct {
	Host string		`yaml:"host"`
	Port string		`yaml:"port"`
}

type EtcdConf struct {
	User string					`yaml:"user"`
	Password string				`yaml:"password"`
	EndPoints []EtcdEndpoint	`yaml:"endPoints"`
}

type ContabLog struct {
	Path string		`yaml:"path"`
}

type CronTabConf struct {
	Etcd EtcdConf		`yaml:"etcdServer"`
	Log ContabLog		`yaml:"logs"`
}

func (cc *CronTabConf) Init() (err error){
	var (
		crontabConfig []byte
	)
	if crontabConfig, err = ioutil.ReadFile("crontab.conf"); err != nil {
		return
	}

	if err = yaml.Unmarshal(crontabConfig, cc); err != nil {
		return
	}

	return
}




