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
	Tiktok  TiktokConfig        `yaml:"tiktok"`
}

type TiktokConfig struct {
	TokenURL                    string                      `yaml:"token_url"`
	USTokenURL                  string                      `yaml:"us_token_url"`
	NonUSTokenURL               string                      `yaml:"non_us_token_url"`
	BaseURL                     string                      `yaml:"base_url"`
	Version                     string                      `yaml:"version"`
	FullVersion                 string                      `yaml:"full_version"`
	FullCategoryAttrVersion     string                      `yaml:"full_category_attr_version"`
	FullCronVersion             string                      `yaml:"full_cron_version"`
	FullListVersion             string                      `yaml:"full_list_version"`
	FullStockUpOrderVersion     string                      `yaml:"full_stock_up_order_version"`
	FullStockUpOrderConfVersion string                      `yaml:"full_stock_up_order_conf_version"`
	AppKey                      string                      `yaml:"app_key"`
	AppSecret                   string                      `yaml:"app_secret"`
	ServiceID                   string                      `yaml:"service_id"`
	AccessToken                 string                      `yaml:"access_token"`
	RefreshToken                string                      `yaml:"refresh_token"`
	SellerType                  string                      `yaml:"seller_type"`
	Region                      string                      `yaml:"region"`
	ShopCipher                  string                      `yaml:"shop_cipher"`
	OpenID                      string                      `yaml:"open_id"`
	ShopSign                    string                      `yaml:"shop_sign"`
	PlatformTiktokID            string                      `yaml:"platform_tiktok_id"`
	CategoryVersion             string                      `yaml:"category_version"`
	Shops                       map[string]TiktokShopConfig `yaml:"shops"`
}

type TiktokShopConfig struct {
	ShopType                    int    `yaml:"shop_type"`
	TokenURL                    string `yaml:"token_url"`
	USTokenURL                  string `yaml:"us_token_url"`
	NonUSTokenURL               string `yaml:"non_us_token_url"`
	BaseURL                     string `yaml:"base_url"`
	Version                     string `yaml:"version"`
	FullVersion                 string `yaml:"full_version"`
	FullCategoryAttrVersion     string `yaml:"full_category_attr_version"`
	FullCronVersion             string `yaml:"full_cron_version"`
	FullListVersion             string `yaml:"full_list_version"`
	FullStockUpOrderVersion     string `yaml:"full_stock_up_order_version"`
	FullStockUpOrderConfVersion string `yaml:"full_stock_up_order_conf_version"`
	AppKey                      string `yaml:"app_key"`
	AppSecret                   string `yaml:"app_secret"`
	ServiceID                   string `yaml:"service_id"`
	AccessToken                 string `yaml:"access_token"`
	RefreshToken                string `yaml:"refresh_token"`
	SellerType                  string `yaml:"seller_type"`
	Region                      string `yaml:"region"`
	ShopCipher                  string `yaml:"shop_cipher"`
	OpenID                      string `yaml:"open_id"`
	ShopSign                    string `yaml:"shop_sign"`
	PlatformTiktokID            string `yaml:"platform_tiktok_id"`
	CategoryVersion             string `yaml:"category_version"`
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

func GetTiktok() TiktokConfig {
	return config.Tiktok
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
