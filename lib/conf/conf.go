package conf

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var LogLevelMap map[string]int

type Conf struct {
	Version string              `yaml:"version"`
	Host    map[string]string   `yaml:"host"`
	Hosts   map[string][]string `yaml:"hosts"`
	Mysql   map[string]string   `yaml:"mysql"`
	Redis   map[string]string   `yaml:"redis"`
	Log     map[string]string   `yaml:"log"`
	File    map[string]string   `yaml:"file"`
	Im      map[string]string   `yaml:"im"`
	Mongo   map[string]string   `yaml:"mongo"`
	AppId   string              `yaml:"app_id"`
}

var config Conf

func GetMongo(name string) string {
	return config.Mongo[name]
}

func GetFile(name string) string {
	return config.File[name]
}

func GetVersion() string {
	return config.Version
}

func GetMysqlSource(name string) string {
	return config.Mysql[name]
}

func GetRedis(name string) string {
	return config.Redis[name]
}

func GetIm(name string) string {
	return config.Im[name]
}

func GetHost(name string) string {
	return config.Host[name]
}

func GetHosts(name string) []string {
	return config.Hosts[name]
}

func GetLog(name string) string {
	return config.Log[name]
}

func GetLogLevel() int {
	return LogLevelMap[GetLog("level")]
}

func GetAppId() string {
	return config.AppId
}

func Init(servername string) {
	prefixPath := os.Getenv("statistics-server-PATH")
	if prefixPath == "" {
		//gopath := os.Getenv("GOPATH")
		prefixPath = "/usr/local/statistics/" + servername + "/etc"
	}
	yamlFile, err := ioutil.ReadFile(fmt.Sprintf("%s/server.yaml", prefixPath))
	if err != nil {
		fmt.Println("read yaml config error: ", err)
	}
	err = yaml.UnmarshalStrict(yamlFile, &config)

	LogLevelMap = map[string]int{"panic": 0, "fatal": 1, "error": 2, "warn": 3, "info": 4, "debug": 5}
}
