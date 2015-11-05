package orders

import (
	"encoding/json"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"strconv"
	"time"
	"utils/alipay"
	"utils/goods"
	"utils/orders"
	. "utils/types"
	//"reflect"
)

func PostOrder(r render.Render, params martini.Params, dbmap *gorp.DbMap, res http.ResponseWriter, o Order, e binding.Errors) {
	if e != nil {
		//r.JSON(http.StatusBadRequest, map[string]string{"message": e[0].Message})
		res.WriteHeader(http.StatusBadRequest)
		res.Header().Set("Content-Type", "application/text")
		res.Write([]byte(e[0].Message))
		return
	}

	alipaytype := params["type"]

	//check user exists yet or not?
	var count int64
	//count,err := dbmap.SelectInt("SELECT count(*) FROM ORDERS WHERE O_ID=?",o.Id)
	row, err := dbmap.Db.Query("SELECT count(*) FROM ORDERS WHERE O_ID=?", o.Id)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by id:%v fail:%v", o.Id, err)
		//r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Header().Set("Content-Type", "application/text")
		res.Write([]byte("DB ERROR"))
		return
	}
	defer row.Close()

	row.Next()
	err = row.Scan(&count)
	if err != nil {
		glog.V(2).Infof("Scan fail:", err.Error())
	}
	//glog.V(2).Infof("Count = %v", count)

	if count > 0 {
		glog.V(1).Infof("Customer with tel:%v exists yet", o.Id)
		//r.JSON(http.StatusConflict, map[string]string{"message": "Order with same Id exists"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Header().Set("Content-Type", "application/text")
		res.Write([]byte("Order with same Id exists"))
		return
	}

	o.Status = 2
	o.CreateTime = NullTime{Time: time.Now(), Valid: true}

	fee, err := goods.GetGoodFeeById(o.GoodId.String, dbmap)
	if err != nil {
		o.Status = 9
		_, err2 := dbmap.Update(&o)
		if err2 != nil {
			glog.V(1).Infof("[DEBUG:] Update Order %v to invalid fail:%v", o, err2)
		}
		//r.JSON(http.StatusBadRequest, map[string]string{"message": "DB ERROR"})
		res.WriteHeader(http.StatusBadRequest)
		res.Header().Set("Content-Type", "application/text")
		res.Write([]byte("Invalid Good ID"))
		return
	}

	o.Sum = fee * float32(o.GoodCnt)

	//insert new user info to db;
	err = dbmap.Insert(&o)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Insert Order %v fail:%v", o, err)
		//r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		res.WriteHeader(http.StatusInternalServerError)
		res.Header().Set("Content-Type", "application/text")
		res.Write([]byte("Order Insert DB Fail"))
		return
	}

	//after saved to db,we call alipay and submit to alipay
	outhtml, outscript := alipay.Form(o, alipaytype)

	ob := map[string]string{"html": outhtml, "script": outscript}
	ob["self"] = "admin"
	res.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(ob)

	//r.JSON(200, outscript)
	res.WriteHeader(http.StatusOK)
	res.Header().Set("Content-Type", "application/text")
	res.Write([]byte(js))
}

func GetOrder(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	id := params["id"]
	stel := params["tel"]
	tel, _ := strconv.ParseInt(stel, 10, 64)
	//get userinfo
	o, err := orders.GetOrderByIdAndTel(id, tel, dbmap)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by tel:%v & id:%v fail:%v", tel, id, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "Order not exists"})
		return
	}

	r.JSON(200, o)
}

func GetOrders(r render.Render, dbmap *gorp.DbMap, res http.ResponseWriter) {
	o, err := orders.GetOrders(dbmap)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by fail:%v", err)
		r.JSON(http.StatusConflict, map[string]string{"message": "Order not exists"})
		return
	}
	r.JSON(200, o)
}

func GetOrdersByStatus(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	status := params["status"]
	var s int64
	if status == "Complete" {
		s = 6
	} else if status == "Paysuccess" {
		s = 5
	} else if status == "Payfail" {
		s = 4
	} else if status == "Shipping" {
		s = 3
	} else {
		r.JSON(http.StatusConflict, map[string]string{"message": "Not support status"})
		return
	}
	o, err := orders.GetOrdersByStatus(s, dbmap)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by status %v fail:%v", status, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "Order not exists"})
		return
	}
	r.JSON(200, o)
}

func GetOrdersByTel(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	tels := params["Tel"]
	tel, err := strconv.ParseInt(tels, 10, 64)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]string{"message": "tel invalid"})
		return
	}
	o, err := orders.GetOrdersByTel(tel, dbmap)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by tel %v fail:%v", tel, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "Order not exists"})
		return
	}
	r.JSON(200, o)
}

func PutOrderStatus(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	orderid := params["id"]
	newstatuss := params["newstatus"]
	newstatus, err := strconv.ParseInt(newstatuss, 10, 64)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]string{"message": "newstatus invalid"})
		return
	}

	oldtatuss := params["oldstatus"]
	oldstatus, err := strconv.ParseInt(oldtatuss, 10, 64)
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]string{"message": "oldstatus invalid"})
		return
	}

	o, err := orders.UpdateOrderFromStatusToStatus(orderid, oldstatus, newstatus, dbmap)
	if err != nil {
		r.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}
	r.JSON(http.StatusOK, o)
}

func PutOrderShipping(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	orderid := params["id"]
	shippingNo := params["shippingNo"]
	shippingCom := params["shippingCom"]

	o, err := orders.UpdateOrderShipping(orderid, shippingNo, shippingCom, dbmap)
	if err != nil {
		r.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}
	r.JSON(http.StatusOK, o)
}

func GetOrderShipping(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	orderid := params["id"]

	o, err := orders.GetOrderById(orderid, dbmap)
	if err != nil {
		r.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}

	//HaoService api
	url := "http://apis.haoservice.com/lifeservice/exp?com=" + o.ShipCom.String + "&no=" + o.ShipNo.String + "&key=" + "798a8f4544184187a48eb9169d5d1ba5"
	resp, body, errs := gorequest.New().Get(url).EndBytes()
	if errs != nil {
		r.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		return
	}

	if resp.StatusCode != 200 {
		r.JSON(http.StatusInternalServerError, map[string]string{"badRsp": resp.Status})
		return
	}

	var content ShipInfo
	_ = json.Unmarshal(body, &content)

	r.JSON(http.StatusOK, content)
}

func Test(r render.Render) {
	CheckOrderStatus()
	r.Status(http.StatusOK)
}

func CheckOrderStatus() {
	os, err := orders.GetOrdersByStatus(3, DbMap)
	if err != nil {
		glog.V(1).Infof("Get Order with PaySuccess fail:%#v", err.Error())
		return
	}
	for _, o := range os {
		//HaoService api
		url := "http://apis.haoservice.com/lifeservice/exp?com=" + o.ShipCom.String + "&no=" + o.ShipNo.String + "&key=" + "798a8f4544184187a48eb9169d5d1ba5"
		resp, body, errs := gorequest.New().Get(url).EndBytes()
		if errs != nil {
			glog.V(1).Infof("Get Order Shipping info fail:%#v", err.Error())
		}

		if resp.StatusCode != 200 {
			glog.V(1).Infof("Get Order Shipping info fail:%#v", err.Error())
		}

		var content ShipInfo
		_ = json.Unmarshal(body, &content)

		if content.Result.IsCheck {
			_, err = orders.UpdateOrderFromStatusToStatus(o.Id.String, 3, 6, DbMap)
			if err != nil {
				glog.V(1).Infof("Update Order %v from Shipping to Completed fail:%#v", err.Error())
			}
		}
	}
}
