package alipay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/golang/glog"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	. "utils/types"
)

// 按照支付宝规则生成sign
func sign_struct(param interface{}) string {
	//解析为字节数组
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return ""
	}

	//重组字符串
	var sign string
	oldString := string(paramBytes)

	//为保证签名前特殊字符串没有被转码，这里解码一次
	oldString = strings.Replace(oldString, `\u003c`, "<", -1)
	oldString = strings.Replace(oldString, `\u003e`, ">", -1)

	//去除特殊标点
	oldString = strings.Replace(oldString, "\"", "", -1)
	oldString = strings.Replace(oldString, "{", "", -1)
	oldString = strings.Replace(oldString, "}", "", -1)
	paramArray := strings.Split(oldString, ",")

	for _, v := range paramArray {
		detail := strings.SplitN(v, ":", 2)
		//排除sign和sign_type
		if detail[0] != "sign" && detail[0] != "sign_type" {
			//total_fee转化为2位小数
			if detail[0] == "total_fee" {
				number, _ := strconv.ParseFloat(detail[1], 32)
				detail[1] = strconv.FormatFloat(number, 'f', 2, 64)
			}
			if sign == "" {
				sign = detail[0] + "=" + detail[1]
			} else {
				sign += "&" + detail[0] + "=" + detail[1]
			}
		}

	}

	//追加密钥
	sign += Config.Alipay.Key //AlipayKey

	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign
}

func sign_raw(raw string) string {
	//重组字符串
	var sign string
	oldString, err := url.QueryUnescape(raw)
	if err != nil {
		return ""
	}

	//为保证签名前特殊字符串没有被转码，这里解码一次
	oldString = strings.Replace(oldString, `\u003c`, "<", -1)
	oldString = strings.Replace(oldString, `\u003e`, ">", -1)

	//去除特殊标点
	oldString = strings.Replace(oldString, "\"", "", -1)
	oldString = strings.Replace(oldString, "{", "", -1)
	oldString = strings.Replace(oldString, "}", "", -1)

	paramArray := strings.Split(oldString, "&")

	glog.V(1).Infoln("----------")
	for _, v := range paramArray {
		detail := strings.SplitN(v, "=", 2)
		//排除sign和sign_type
		if detail[0] != "sign" && detail[0] != "sign_type" {
			//total_fee转化为2位小数
			if detail[0] == "total_fee" {
				number, _ := strconv.ParseFloat(detail[1], 32)
				detail[1] = strconv.FormatFloat(number, 'f', 2, 64)
			}
			if sign == "" {
				sign = detail[0] + "=" + detail[1]
			} else {
				sign += "&" + detail[0] + "=" + detail[1]
			}
		}
		glog.V(2).Infof("%s ==> %s", detail[0], detail[1])

	}
	glog.V(1).Infoln("----------")

	//追加密钥
	sign += Config.Alipay.Key //AlipayKey

	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign
}

/* 生成支付宝即时到帐提交表单html代码 */
func buildRequest_WEBBank(o Order, defaultBank string) (string, string) {
	//实例化参数
	param := &AlipayParametersWebBank{}
	param.InputCharset = "utf-8"
	param.NotifyUrl = Config.Alipay.NotifyUrl //WebNotifyUrl
	param.OutTradeNo = o.Id.String
	param.Partner = Config.Alipay.Partner //AlipayPartner
	param.PaymentType = 1
	param.ReturnUrl = Config.Alipay.ReturnUrl //WebReturnUrl
	param.SellerEmail = Config.Alipay.Email   //WebSellerEmail
	param.Service = "create_direct_pay_by_user"
	param.Defaultbank = defaultBank
	param.Paymethod = "bankPay"

	param.Subject = "购买妈妈乐付账"
	param.TotalFee = o.Sum

	//生成签名
	sign := sign_struct(param)

	//追加参数
	param.Sign = sign
	param.SignType = "MD5"

	//生成自动提交form并且直接GET到支付宝
	return `
		<form id="alipaysubmit" name="alipaysubmit" action="https://mapi.alipay.com/gateway.do?_input_charset=utf-8" method="get" style='display:none;'>
			<input type="hidden" name="_input_charset" value="` + param.InputCharset + `">
			<input type="hidden" name="defaultbank" value="` + param.Defaultbank + `">
			<input type="hidden" name="notify_url" value="` + param.NotifyUrl + `">
			<input type="hidden" name="out_trade_no" value="` + param.OutTradeNo + `">
			<input type="hidden" name="partner" value="` + param.Partner + `">
			<input type="hidden" name="payment_type" value="` + strconv.Itoa(int(param.PaymentType)) + `">
			<input type="hidden" name="paymethod" value="` + param.Paymethod + `">
			<input type="hidden" name="return_url" value="` + param.ReturnUrl + `">
			<input type="hidden" name="seller_email" value="` + param.SellerEmail + `">
			<input type="hidden" name="service" value="` + param.Service + `">
			<input type="hidden" name="subject" value="` + param.Subject + `">
			<input type="hidden" name="total_fee" value="` + strconv.FormatFloat(float64(param.TotalFee), 'f', 2, 32) + `">
			<input type="hidden" name="sign" value="` + param.Sign + `">
			<input type="hidden" name="sign_type" value="` + param.SignType + `">
		</form>`,
		`
			document.forms['alipaysubmit'].submit();
		`
}

func buildRequest_WEB(o Order) (string, string) {
	//实例化参数
	param := &AlipayParametersWeb{}
	param.InputCharset = "utf-8"
	param.Body = "测试支付宝接口"
	param.NotifyUrl = Config.Alipay.NotifyUrl //WebNotifyUrl
	param.OutTradeNo = o.Id.String
	param.Partner = Config.Alipay.Partner //AlipayPartner
	param.PaymentType = 1
	param.ReturnUrl = Config.Alipay.ReturnUrl //WebReturnUrl
	param.SellerEmail = Config.Alipay.Email   //WebSellerEmail
	param.Service = "create_direct_pay_by_user"

	param.Subject = "购买妈妈乐付账"
	param.TotalFee = o.Sum

	//生成签名
	sign := sign_struct(param)

	//追加参数
	param.Sign = sign
	param.SignType = "MD5"

	//生成自动提交form并且直接GET到支付宝
	return `
		<form id="alipaysubmit" name="alipaysubmit" action="https://mapi.alipay.com/gateway.do?_input_charset=utf-8" method="get" style='display:none;'>
			<input type="hidden" name="_input_charset" value="` + param.InputCharset + `">
			<input type="hidden" name="body" value="` + param.Body + `">
			<input type="hidden" name="notify_url" value="` + param.NotifyUrl + `">
			<input type="hidden" name="out_trade_no" value="` + param.OutTradeNo + `">
			<input type="hidden" name="partner" value="` + param.Partner + `">
			<input type="hidden" name="payment_type" value="` + strconv.Itoa(int(param.PaymentType)) + `">
			<input type="hidden" name="return_url" value="` + param.ReturnUrl + `">
			<input type="hidden" name="seller_email" value="` + param.SellerEmail + `">
			<input type="hidden" name="service" value="` + param.Service + `">
			<input type="hidden" name="subject" value="` + param.Subject + `">
			<input type="hidden" name="total_fee" value="` + strconv.FormatFloat(float64(param.TotalFee), 'f', 2, 32) + `">
			<input type="hidden" name="sign" value="` + param.Sign + `">
			<input type="hidden" name="sign_type" value="` + param.SignType + `">
		</form>`,
		`
			document.forms['alipaysubmit'].submit();
		`
}

func buildRequest_WAP(o Order) (string, string) {
	//实例化参数
	param := &AlipayParametersWap{}
	param.InputCharset = "utf-8"
	param.Body = "测试支付宝接口"
	param.NotifyUrl = Config.Alipay.NotifyUrl //WebNotifyUrl
	param.OutTradeNo = o.Id.String
	param.Partner = Config.Alipay.Partner //AlipayPartner
	param.PaymentType = 1
	param.ReturnUrl = Config.Alipay.ReturnUrl //WebReturnUrl
	param.SellerId = Config.Alipay.Partner    //WebSellerId
	param.Service = "alipay.wap.create.direct.pay.by.user"

	param.Subject = "购买妈妈乐付账"
	param.TotalFee = o.Sum

	//生成签名
	sign := sign_struct(param)

	//追加参数
	param.Sign = sign
	param.SignType = "MD5"

	//生成自动提交form并且直接GET到支付宝
	return `
		<form id="alipaysubmit" name="alipaysubmit" action="https://mapi.alipay.com/gateway.do?_input_charset=utf-8" method="get" style='display:none;'>
			<input type="hidden" name="_input_charset" value="` + param.InputCharset + `">
			<input type="hidden" name="body" value="` + param.Body + `">
			<input type="hidden" name="notify_url" value="` + param.NotifyUrl + `">
			<input type="hidden" name="out_trade_no" value="` + param.OutTradeNo + `">
			<input type="hidden" name="partner" value="` + param.Partner + `">
			<input type="hidden" name="payment_type" value="` + strconv.Itoa(int(param.PaymentType)) + `">
			<input type="hidden" name="return_url" value="` + param.ReturnUrl + `">
			<input type="hidden" name="seller_id" value="` + param.SellerId + `">
			<input type="hidden" name="service" value="` + param.Service + `">
			<input type="hidden" name="subject" value="` + param.Subject + `">
			<input type="hidden" name="total_fee" value="` + strconv.FormatFloat(float64(param.TotalFee), 'f', 2, 32) + `">
			<input type="hidden" name="sign" value="` + param.Sign + `">
			<input type="hidden" name="sign_type" value="` + param.SignType + `">
		</form>`,
		`
			document.forms['alipaysubmit'].submit();
		`
}

func Form(o Order, alipay string) (string, string) {
	if alipay == "1" {
		return buildRequest_WEB(o)
	} else if alipay == "2" {
		return buildRequest_WAP(o)
	} else {
		return buildRequest_WEBBank(o, alipay)
	}
}

/* 被动接收支付宝同步跳转的页面 */
func Return(params map[string]string, raw string) *Result {
	//var paramweb AlipayNotifyParamsWeb
	//var paramwap AlipayNotifyParamsWap
	//var parambank AlipayNotifyParamsBank
	//var isWeb bool
	// 结果
	result := &Result{}

	//martini.Params -> AlipayNotifyParams
	//	m, err := json.Marshal(params)
	//	if params["service"] == "alipay.wap.create.direct.pay.by.user" {
	//		err = json.Unmarshal(m, &paramwap)
	//		isWeb = false
	//	} else {
	//		err = json.Unmarshal(m, &paramweb)
	//		isWeb = true
	//	}

	//	 解析表单内容，失败返回错误代码-3
	//	if err != nil {
	//		result.Status = -3
	//		result.Message = "解析表单失败"
	//		glog.V(1).Infof("[DEBUG:] Return %v:%v", result.Message, err)
	//		return result
	//	}

	//	glog.V(1).Infoln("----------")
	//	glog.V(2).Infof("参数：%#v", params)
	//	if isWeb {
	//		glog.V(2).Infof("WEB结构：%#v", paramweb)
	//	} else {
	//		glog.V(2).Infof("WAP结构：%#v", paramwap)
	//	}
	//	glog.V(1).Infoln("----------")

	// 如果最基本的网站交易号为空，返回错误代码-1
	if params["out_trade_no"] == "" { //不存在交易号
		result.Status = -1
		result.Message = "站交易号为空"
		glog.V(1).Infof("[DEBUG:] Return %v", result.Message)
		return result
	} else {
		// 生成签名
		var sig string

		//		if isWeb {
		//			sig = sign(paramweb)
		//		} else {
		//			sig = sign(paramwap)
		//		}
		sig = sign_raw(raw)

		// 对比签名是否相同
		if sig == params["sign"] { //只有相同才说明该订单成功了
			// 判断订单是否已完成
			if params["trade_status"] == "TRADE_FINISHED" || params["trade_status"] == "TRADE_SUCCESS" { //交易成功
				result.Status = 1
				result.OrderNo = params["out_trade_no"]
				result.TradeNo = params["trade_no"]
				result.BuyerEmail = params["buyer_email"]
				return result
			} else { // 交易未完成，返回错误代码-4
				result.Status = -4
				result.Message = "交易未完成"
				glog.V(1).Infof("[DEBUG:] Return %v", result.Message)
				return result
			}
		} else { // 签名认证失败，返回错误代码-2
			result.Status = -2
			result.Message = "签名认证失败"
			glog.V(1).Infof("[DEBUG:] Return %v", result.Message)
			return result
		}
	}

	// 位置错误类型-5
	result.Status = -5
	result.Message = "位置错误"
	glog.V(1).Infof("[DEBUG:] Return %v", result.Message)
	return result
}

/* 被动接收支付宝异步通知 */
func Notify(body string) *Result {
	// 从body里读取参数，用&切割
	postArray := strings.Split(body, "&")

	// 实例化url
	urls := &url.Values{}

	// 保存传参的sign
	var paramSign string
	var sign string

	// 如果字符串中包含sec_id说明是手机端的异步通知
	if strings.Index(body, `alipay.wap.trade.create.direct`) == -1 { // 快捷支付
		for _, v := range postArray {
			detail := strings.Split(v, "=")

			// 使用=切割字符串 去除sign和sign_type
			if detail[0] == "sign" || detail[0] == "sign_type" {
				if detail[0] == "sign" {
					paramSign = detail[1]
				}
				continue
			} else {
				urls.Add(detail[0], detail[1])
			}
		}

		// url解码
		urlDecode, _ := url.QueryUnescape(urls.Encode())
		sign, _ = url.QueryUnescape(urlDecode)
	} else { // 手机网页支付
		// 手机字符串加密顺序
		mobileOrder := []string{"service", "v", "sec_id", "notify_data"}
		for _, v := range mobileOrder {
			for _, value := range postArray {
				detail := strings.Split(value, "=")
				// 保存sign
				if detail[0] == "sign" {
					paramSign = detail[1]
				} else {
					// 如果满足当前v
					if detail[0] == v {
						if sign == "" {
							sign = detail[0] + "=" + detail[1]
						} else {
							sign += "&" + detail[0] + "=" + detail[1]
						}
					}
				}
			}
		}
		sign, _ = url.QueryUnescape(sign)

		// 获取<trade_status></trade_status>之间的request_token
		re, _ := regexp.Compile("\\<trade_status[\\S\\s]+?\\</trade_status>")
		rt := re.FindAllString(sign, 1)
		trade_status := strings.Replace(rt[0], "<trade_status>", "", -1)
		trade_status = strings.Replace(trade_status, "</trade_status>", "", -1)
		urls.Add("trade_status", trade_status)

		// 获取<out_trade_no></out_trade_no>之间的request_token
		re, _ = regexp.Compile("\\<out_trade_no[\\S\\s]+?\\</out_trade_no>")
		rt = re.FindAllString(sign, 1)
		out_trade_no := strings.Replace(rt[0], "<out_trade_no>", "", -1)
		out_trade_no = strings.Replace(out_trade_no, "</out_trade_no>", "", -1)
		urls.Add("out_trade_no", out_trade_no)

		// 获取<buyer_email></buyer_email>之间的request_token
		re, _ = regexp.Compile("\\<buyer_email[\\S\\s]+?\\</buyer_email>")
		rt = re.FindAllString(sign, 1)
		buyer_email := strings.Replace(rt[0], "<buyer_email>", "", -1)
		buyer_email = strings.Replace(buyer_email, "</buyer_email>", "", -1)
		urls.Add("buyer_email", buyer_email)

		// 获取<trade_no></trade_no>之间的request_token
		re, _ = regexp.Compile("\\<trade_no[\\S\\s]+?\\</trade_no>")
		rt = re.FindAllString(sign, 1)
		trade_no := strings.Replace(rt[0], "<trade_no>", "", -1)
		trade_no = strings.Replace(trade_no, "</trade_no>", "", -1)
		urls.Add("trade_no", trade_no)
	}
	// 追加密钥
	sign += Config.Alipay.Key // AlipayKey

	// 返回参数
	result := &Result{}

	// md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	if paramSign == sign { // 传进的签名等于计算出的签名，说明请求合法
		// 判断订单是否已完成
		if urls.Get("trade_status") == "TRADE_FINISHED" || urls.Get("trade_status") == "TRADE_SUCCESS" { //交易成功
			//contro.Ctx.WriteString("success")
			result.Status = 1
			result.OrderNo = urls.Get("out_trade_no")
			result.TradeNo = urls.Get("trade_no")
			result.BuyerEmail = urls.Get("buyer_email")
			return result
		} else {
			//contro.Ctx.WriteString("error")
		}
	} else {
		//contro.Ctx.WriteString("error")
		// 签名不符，错误代码-1
		result.Status = -1
		result.Message = "签名不符"
		return result
	}
	// 未知错误-2
	result.Status = -2
	result.Message = "未知错误"
	return result
}
