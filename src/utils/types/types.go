package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type StatusType int64

const (
	New StatusType = iota
	Delete
	Paying
	Shipping
	Payfail
	Paysuccess
	Complete
	Cannel
	Returning
	Invalid
)

func (s *StatusType) Scan(value interface{}) error {
	if value != nil {
		if v, ok := value.(int64); ok {
			*s = StatusType(v)
		} else {
			*s = Invalid
		}
	} else {
		*s = New
	}

	return nil
}

func (s StatusType) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *StatusType) UnmarshalJSON(data []byte) error {
	d, err := strconv.ParseInt(string(data), 10, 64)
	*s = StatusType(d)
	return err
}

func (s StatusType) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(s))
}

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	//nt.Time, nt.Valid = value.(time.Time)
	if value != nil {
		t, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", string(value.([]byte)))
		if err != nil {
			fmt.Printf("Timestamp:%v", err.Error())
			nt.Valid = false
			return err
		}
		nt.Time = t
		nt.Valid = true
	} else {
		nt.Valid = false
	}
	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	layout := "Jan 2, 2006 at 3:04pm (MST)"
	return nt.Time.Format(layout), nil
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	tmp := nt.Time
	err := (&tmp).UnmarshalJSON(data)
	nt.Time = tmp
	return err
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	return nt.Time.MarshalJSON()
}

type NullString struct {
	sql.NullString
}

func (s *NullString) UnmarshalJSON(data []byte) error {
	s.String = strings.Trim(string(data), `"`)
	//fmt.Printf("JsonValue:%v",string(data))
	//s.String = string(data)
	s.Valid = true
	return nil
}

func (s NullString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String)
}

type Order struct {
	Id         NullString `json:"id" db:"O_ID"`
	GoodId     NullString `json:"goodId" db:"O_GID"`
	GoodCnt    int64      `json:"goodcount" db:"O_GCNT"`
	Desc       NullString `json:"desc" db:"O_DESC"`
	CusName    NullString `json:"cusname" db:"O_CUS_N"`
	CusTel     int64      `json:"custel" db:"O_CUS_TEL"`
	CusAddr    NullString `json:"cusaddr" db:"O_CUS_ADDR"`
	Status     int64      `json:"status" db:"O_ST"`
	CreateTime NullTime   `json:"createtime" db:"O_CT"`
	PayTime    NullTime   `json:"paytime" db:"O_PT"`
	Sum        float32    `json:"sum" db:"O_SUM"`
	ShipNo     NullString `db:"O_SHIP_NO"`
	ShipCom    NullString `db:"O_SHIP_COM"`
}

func (o Order) Validate(errors binding.Errors, req *http.Request) binding.Errors {
	if len(strconv.FormatInt(o.CusTel, 10)) < 11 {
		glog.V(2).Infof("Tel %v is too short!", o.CusTel)
		//errors.Fields["title"] = "Too short, minimum 11 characters"
		errors = append(errors, binding.Error{
			FieldNames:     []string{"message"},
			Classification: "ComplaintError",
			Message:        "Tel is too short!",
		})
		return errors
	}
	if o.GoodCnt < 1 {
		glog.V(2).Infof("Good Count %v Invalid!", o.GoodCnt)
		//errors.Fields["title"] = "Too short, minimum 11 characters"
		errors = append(errors, binding.Error{
			FieldNames:     []string{"message"},
			Classification: "ComplaintError",
			Message:        "Good Count is invalid!",
		})
		return errors
	}

	return nil
}

type Result struct {
	// 状态
	Status int
	// 本网站订单号
	OrderNo string
	// 支付宝交易号
	TradeNo string
	// 买家支付宝账号
	BuyerEmail string
	// 错误提示
	Message string
}

// 生成订单的参数
type Options struct {
	OrderId  string  // 订单唯一id
	Fee      float32 // 价格
	NickName string  // 充值账户名称
	Subject  string  // 充值描述
}

type AlipayParametersWeb struct {
	InputCharset string  `json:"_input_charset"` //网站编码
	Body         string  `json:"body"`           //订单描述
	NotifyUrl    string  `json:"notify_url"`     //异步通知页面
	OutTradeNo   string  `json:"out_trade_no"`   //订单唯一id
	Partner      string  `json:"partner"`        //合作者身份ID
	PaymentType  uint8   `json:"payment_type"`   //支付类型 1：商品购买
	ReturnUrl    string  `json:"return_url"`     //回调url
	SellerEmail  string  `json:"seller_email"`   //卖家支付宝邮箱
	Service      string  `json:"service"`        //接口名称
	Subject      string  `json:"subject"`        //商品名称
	TotalFee     float32 `json:"total_fee"`      //总价
	Sign         string  `json:"sign"`           //签名，生成签名时忽略
	SignType     string  `json:"sign_type"`      //签名类型，生成签名时忽略
}

type AlipayParametersWap struct {
	InputCharset string  `json:"_input_charset"` //网站编码
	Body         string  `json:"body"`           //订单描述
	NotifyUrl    string  `json:"notify_url"`     //异步通知页面
	OutTradeNo   string  `json:"out_trade_no"`   //订单唯一id
	Partner      string  `json:"partner"`        //合作者身份ID
	PaymentType  uint8   `json:"payment_type"`   //支付类型 1：商品购买
	ReturnUrl    string  `json:"return_url"`     //回调url
	SellerId     string  `json:"seller_id"`      //卖家支付宝唯一用户号
	Service      string  `json:"service"`        //接口名称
	Subject      string  `json:"subject"`        //商品名称
	TotalFee     float32 `json:"total_fee"`      //总价
	Sign         string  `json:"sign"`           //签名，生成签名时忽略
	SignType     string  `json:"sign_type"`      //签名类型，生成签名时忽略
}

type AlipayParametersWebBank struct {
	InputCharset string  `json:"_input_charset"` //网站编码
	Defaultbank  string  `json:"defaultbank"`    //默认网银
	NotifyUrl    string  `json:"notify_url"`     //异步通知页面
	OutTradeNo   string  `json:"out_trade_no"`   //订单唯一id
	Partner      string  `json:"partner"`        //合作者身份ID
	PaymentType  uint8   `json:"payment_type"`   //支付类型 1：商品购买
	Paymethod    string  `json:"paymethod"`      //默认支付方式
	ReturnUrl    string  `json:"return_url"`     //回调url
	SellerEmail  string  `json:"seller_email"`   //卖家支付宝邮箱
	Service      string  `json:"service"`        //接口名称
	Subject      string  `json:"subject"`        //商品名称
	TotalFee     float32 `json:"total_fee"`      //总价
	Sign         string  `json:"sign"`           //签名，生成签名时忽略
	SignType     string  `json:"sign_type"`      //签名类型，生成签名时忽略
}

// 列举全部传参
type AlipayNotifyParamsWeb struct {
	Body        string `form:"body" json:"body"`                 // 描述
	BuyerEmail  string `form:"buyer_email" json:"buyer_email"`   // 买家账号
	BuyerId     string `form:"buyer_id" json:"buyer_id"`         // 买家ID
	Exterface   string `form:"exterface" json:"exterface"`       // 接口名称
	IsSuccess   string `form:"is_success" json:"is_success"`     // 交易是否成功
	NotifyId    string `form:"notify_id" json:"notify_id"`       // 通知校验id
	NotifyTime  string `form:"notify_time" json:"notify_time"`   // 校验时间
	NotifyType  string `form:"notify_type" json:"notify_type"`   // 校验类型
	OutTradeNo  string `form:"out_trade_no" json:"out_trade_no"` // 在网站中唯一id
	PaymentType string `form:"payment_type" json:"payment_type"` // 支付类型
	SellerEmail string `form:"seller_email" json:"seller_email"` // 卖家账号
	SellerId    string `form:"seller_id" json:"seller_id"`       // 卖家id
	Subject     string `form:"subject" json:"subject"`           // 商品名称
	TotalFee    string `form:"total_fee" json:"total_fee"`       // 总价
	TradeNo     string `form:"trade_no" json:"trade_no"`         // 支付宝交易号
	TradeStatus string `form:"trade_status" json:"trade_status"` // 交易状态 TRADE_FINISHED或TRADE_SUCCESS表示交易成功
	Sign        string `form:"sign" json:"sign"`                 // 签名
	SignType    string `form:"sign_type" json:"sign_type"`       // 签名类型
}

type AlipayNotifyParamsWap struct {
	Body        string `form:"body" json:"body"`                 // 描述
	IsSuccess   string `form:"is_success" json:"is_success"`     // 交易是否成功
	NotifyId    string `form:"notify_id" json:"notify_id"`       // 通知校验id
	NotifyTime  string `form:"notify_time" json:"notify_time"`   // 校验时间
	NotifyType  string `form:"notify_type" json:"notify_type"`   // 校验类型
	OutTradeNo  string `form:"out_trade_no" json:"out_trade_no"` // 在网站中唯一id
	PaymentType string `form:"payment_type" json:"payment_type"` // 支付类型
	SellerId    string `form:"seller_id" json:"seller_id"`       // 卖家id
	Service     string `form:"service" json:"service"`           // 接口名称
	Subject     string `form:"subject" json:"subject"`           // 商品名称
	TotalFee    string `form:"total_fee" json:"total_fee"`       // 总价
	TradeNo     string `form:"trade_no" json:"trade_no"`         // 支付宝交易号
	TradeStatus string `form:"trade_status" json:"trade_status"` // 交易状态 TRADE_FINISHED或TRADE_SUCCESS表示交易成功
	Sign        string `form:"sign" json:"sign"`                 // 签名
	SignType    string `form:"sign_type" json:"sign_type"`       // 签名类型
}

type AlipayNotifyParamsBank struct {
	IsSuccess        string `form:"is_success" json:"is_success"`                 // 交易是否成功
	Sign             string `form:"sign" json:"sign"`                             // 签名
	SignType         string `form:"sign_type" json:"sign_type"`                   // 签名类型
	Body             string `form:"body" json:"body"`                             // 描述
	BuyerEmail       string `form:"buyer_email" json:"buyer_email"`               // 买家账号
	BuyerId          string `form:"buyer_id" json:"buyer_id"`                     // 买家ID
	Exterface        string `form:"exterface" json:"exterface"`                   // 接口名称
	OutTradeNo       string `form:"out_trade_no" json:"out_trade_no"`             // 在网站中唯一id
	PaymentType      string `form:"payment_type" json:"payment_type"`             // 支付类型
	SellerEmail      string `form:"seller_email" json:"seller_email"`             // 卖家账号
	SellerId         string `form:"seller_id" json:"seller_id"`                   // 卖家id
	Subject          string `form:"subject" json:"subject"`                       // 商品名称
	TotalFee         string `form:"total_fee" json:"total_fee"`                   // 总价
	TradeNo          string `form:"trade_no" json:"trade_no"`                     // 支付宝交易号
	TradeStatus      string `form:"trade_status" json:"trade_status"`             // 交易状态 TRADE_FINISHED或TRADE_SUCCESS表示交易成功
	NotifyId         string `form:"notify_id" json:"notify_id"`                   // 通知校验id
	NotifyTime       string `form:"notify_time" json:"notify_time"`               // 校验时间
	NotifyType       string `form:"notify_type" json:"notify_type"`               // 校验类型
	ExtraCommonParam string `form:"extra_common_param" json:"extra_common_param"` //公用回传参数
	BankSeqNo        string `form:"bank_seq_no" json:"bank_seq_no"`               //网银流水
}

type User struct {
	PhoneNumber int64  `form:"phonenumber" json:"phonenumber"  db:"CUS_TEL" binding:"required"`
	Email       string `form:"email" json:"email" db:"CUS_EMAIL"`
	Password    string `form:"password" json:"password" db:"CUS_PW" binding:"required"`
	Name        string `form:"name" json:"name" db:"CUS_NAME"  binding:"required"`
	Address     string `form:"address" json:"address" db:"CUS_ADDR" binding:"required"`
}

func (u User) Validate(errors binding.Errors, req *http.Request) binding.Errors {
	if len(strconv.FormatInt(u.PhoneNumber, 10)) < 11 {
		glog.V(2).Infof("Tel %v is too short!", u.PhoneNumber)
		//errors.Fields["title"] = "Too short, minimum 11 characters"
		errors = append(errors, binding.Error{
			FieldNames:     []string{"message"},
			Classification: "ComplaintError",
			Message:        "Tel is too short!",
		})
		return errors
	}
	return nil
}

type GoodInfo struct {
	Id          NullString `json:"id" db:"G_ID"`
	Fee         float32    `json:"fee" db:"G_FEE"`
	Description NullString `json:"desc" db:"G_DESC"`
	CreateTime  NullTime   `json:"createtime" db:"G_CT"`
	UpdateTime  NullTime   `json:"updatetime" db:"G_UT"`
}

func (g GoodInfo) Validate(errors binding.Errors, req *http.Request) binding.Errors {
	if g.Fee < 0 {
		glog.V(2).Infof("Fee %v should be positive value!", g.Fee)
		//errors.Fields["title"] = "Too short, minimum 11 characters"
		errors = append(errors, binding.Error{
			FieldNames:     []string{"message"},
			Classification: "ComplaintError",
			Message:        "Fee is invalid!",
		})
		return errors
	}
	return nil
}

type shipContent struct {
	Time    string `json:"time"`
	Context string `json:"context"`
}

type shipResult struct {
	No         string        `json:"no"`
	IsCheck    bool          `json:"ischeck"`
	Com        string        `json:"com"`
	Company    string        `json:"company"`
	UpdateTime string        `json:"updatetime"`
	Data       []shipContent `json:"data" `
}

type ShipInfo struct {
	ErrorCode int        `json:"error_code"`
	Reason    string     `json:"reason"`
	Result    shipResult `json:"result"`
}
