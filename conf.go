package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type DbInfo struct {
	Type     string `yaml:"type"`
	Uri      string `yaml:"uri"`
	User     string `yaml:"user"`
	Pwd      string `yaml:"pwd"`
	DataName string `yaml:"dataname"`
}

type Conf struct {
	Dbinfo       DbInfo `yaml:"dbinfo"`
	GoroutineNum int    `yaml:"goroutinenum"`
	RetryTime    int    `yaml:"retrytime"`
	Limit        int    `yaml:"limit"`
	LogFile      string `yaml:"logfile"`
}

func (c *Conf) ConfReader() {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	err = yaml.Unmarshal([]byte(yamlFile), &c)
	if err != nil {
		log.Fatalf("error:%v", err)
	}
}
