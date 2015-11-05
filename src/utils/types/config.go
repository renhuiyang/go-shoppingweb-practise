package types

import (
	"github.com/coopernurse/gorp"
	)

type Database struct{
	Dbtype string 
	Server string
	Dbname string
	ConnMax int
	User string
	Password string 
	}

type AlipayConfig struct {
	Partner   string // 合作者ID
	Key       string // 合作者私钥
	ReturnUrl string // 同步返回地址
	NotifyUrl string // 网站异步返回地址
	Email     string // 网站卖家邮箱地址
}

type TomlConfig struct{
	DB Database `toml:"database"`
	Alipay AlipayConfig `tomal:"alipay"`
	}

//var (
//	AlipayPartner  string //合作者ID
//	AlipayKey      string //合作者私钥
//	WebReturnUrl   string //网站同步返回地址
//	WebNotifyUrl   string //网站异步返回地址
//	WebSellerEmail string //网站卖家邮箱地址
//)

var Config TomlConfig
var DbMap *gorp.DbMap
