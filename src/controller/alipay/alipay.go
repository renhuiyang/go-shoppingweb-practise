// @authors     ascoders

package alipay

import (
	//"github.com/astaxie/beego"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	//"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"net/http"
	"net/url"
	. "utils/alipay"
	"utils/orders"
	//. "utils/types"
)

func Test(r render.Render, req *http.Request, dbmap *gorp.DbMap, res http.ResponseWriter) {
	//orderId := params["id"]
	qs := req.URL.Query()
	orderId := qs.Get("id")

	m, _ := url.ParseQuery(req.URL.RawQuery)
	params := map[string]string{}
	for k, v := range m {
		params[k] = v[0]
	}
	fmt.Println(params)
	
	type OrderInfo struct {
		Result  bool
		OrderId string
		GoodId  string
		GoodCnt int64
		Tel     int64
		Name    string
		Addr    string
		Sum     float32
	}
	var orderInfo OrderInfo
	orderInfo.Result = false
	o, err := orders.GetOrderById(orderId, dbmap)
	if err != nil {
		r.HTML(400, "orderresult", orderInfo)
		return
	}
	
	orderInfo.Result = true
	orderInfo.OrderId = o.Id.String
	orderInfo.GoodId = o.GoodId.String
	orderInfo.GoodCnt = o.GoodCnt
	orderInfo.Tel = o.CusTel
	orderInfo.Name = o.CusName.String
	orderInfo.Addr = o.CusAddr.String
	orderInfo.Sum = o.Sum

	//r.Status(200)
	r.HTML(200, "orderresult", orderInfo)
}

//process post alipay notify
func TestNotify(r render.Render, req *http.Request, dbmap *gorp.DbMap, res http.ResponseWriter) {
	js, _ := json.Marshal("success")
	res.Write(js)
	return
}


//process post alipay return
func AlipayReturn(r render.Render, req *http.Request, dbmap *gorp.DbMap, res http.ResponseWriter) {
	m, _ := url.ParseQuery(req.URL.RawQuery)
	params := map[string]string{}
	for k, v := range m {
		params[k] = v[0]
	}
	result := Return(params, req.URL.RawQuery)

	type OrderInfo struct {
		Result  bool
		OrderId string
		GoodId  string
		GoodCnt int64
		Tel     int64
		Name    string
		Addr    string
		Sum     float32
	}

	var orderInfo OrderInfo
	orderInfo.Result = false
	if result.Status == -1 || result.Status == -5 || result.Status == -3 {
		r.HTML(400, "orderresult", orderInfo)
		return
	}

	o, err := orders.GetOrderById(result.OrderNo, dbmap)
	if err != nil {
		r.HTML(400, "orderresult", orderInfo)
		return
	}

	if result.Status != 1 {
		orders.UpdateOrderFromStatusToStatus(result.OrderNo, 2, 4, dbmap)
		r.HTML(400, "orderresult", orderInfo)
		return
	}

	_,err = orders.UpdateOrderFromStatusToStatus(result.OrderNo, 2,5, dbmap)
	if err != nil {
		r.HTML(400, "orderresult", orderInfo)
		return
	}

	orderInfo.Result = true
	orderInfo.OrderId = o.Id.String
	orderInfo.GoodId = o.GoodId.String
	orderInfo.GoodCnt = o.GoodCnt
	orderInfo.Tel = o.CusTel
	orderInfo.Name = o.CusName.String
	orderInfo.Addr = o.CusAddr.String
	orderInfo.Sum = o.Sum

	//r.Status(200)
	r.HTML(200, "orderresult", orderInfo)
	return
}

//process post alipay notify
func AlipayNotify(r render.Render, req *http.Request, dbmap *gorp.DbMap, res http.ResponseWriter) {
	// Read the content
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body)
	}
	// Restore the io.ReadCloser to its original state
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	bodyString := string(bodyBytes)

	result := Notify(bodyString)

	if result.Status == -2 {
		//r.Status(400)
		js, _ := json.Marshal("error")
		res.Write(js)
		return
	}

	_, err := orders.GetOrderById(result.OrderNo, dbmap)
	if err != nil {
		js, _ := json.Marshal("error")
		res.Write(js)
		return
	}

	if result.Status != 1 {
		orders.UpdateOrderFromStatusToStatus(result.OrderNo, 2,4, dbmap)
		js, _ := json.Marshal("error")
		res.Write(js)
		return
	}

	_,err = orders.UpdateOrderFromStatusToStatus(result.OrderNo, 2,5, dbmap)
	if err != nil {
		js, _ := json.Marshal("error")
		res.Write(js)
		return
	}
	js, _ := json.Marshal("success")
	res.Write(js)
	return
}
