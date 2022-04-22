package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	configPath      string
	defaultConfiger *Configer
)

func init() {
	configPath = filepath.Join(os.Getenv("HOME"), ".clip/config.json")
	defaultConfiger = NewConfiger(readConfigFile())
}

type Configer struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

func NewConfiger(data []byte) *Configer {
	var config Configer
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
	return &config
}

func readConfigFile() []byte {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return []byte("{}")
	}
	return data
}

func (c *Configer) GetPort() int {
	if c.Port == 0 {
		c.Port = 4321
	}
	return c.Port
}

func (c *Configer) GetHost() string {
	return c.Host
}

func (c *Configer) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.GetHost(), c.GetPort())
}

func GetPort() int    { return defaultConfiger.GetPort() }
func GetHost() string { return defaultConfiger.GetHost() }
func GetAddr() string { return defaultConfiger.GetAddr() }
