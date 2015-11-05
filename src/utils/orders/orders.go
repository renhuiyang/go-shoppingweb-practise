package orders

import (
	"github.com/coopernurse/gorp"
	//"github.com/go-martini/martini"
	"github.com/golang/glog"
	//"github.com/martini-contrib/binding"
	//"github.com/martini-contrib/render"
	//"net/http"
	//"strconv"
	"time"
	. "utils/types"
	//"reflect"
	"fmt"
)

func GetOrderByIdAndTel(oid string, tel int64, dbmap *gorp.DbMap) (Order, error) {
	var o Order
	err := dbmap.SelectOne(&o, "SELECT * FROM ORDERS WHERE O_ID=? and O_CUS_TEL=?", oid, tel)
	//err := db.QueryRow("SELECT CUS_NAME,CUS_ADDR,CUS_EMAIL,CUS_TEL,CUS_PW FROM CUSTOMER WHERE CUS_TEL=?",tel).Scan(&u.Name,&u.Address,&u.Email,&u.PhoneNumber,&u.Password)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by tel:%v & id:%v fail:%v", tel, oid, err)
		return o, err
	}
	return o, nil
}

func GetOrderById(oid string, dbmap *gorp.DbMap) (Order, error) {
	var o Order
	err := dbmap.SelectOne(&o, "SELECT * FROM ORDERS WHERE O_ID=?", oid)
	//err := db.QueryRow("SELECT CUS_NAME,CUS_ADDR,CUS_EMAIL,CUS_TEL,CUS_PW FROM CUSTOMER WHERE CUS_TEL=?",tel).Scan(&u.Name,&u.Address,&u.Email,&u.PhoneNumber,&u.Password)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by id:%v fail:%v", oid, err)
		return o, err
	}
	return o, nil
}

func GetOrders(dbmap *gorp.DbMap) ([]Order, error) {
	var os []Order

	//_, err := dbmap.Select(&os, "SELECT * FROM ORDERS WHERE O_ID=?")
	rows, err := dbmap.Db.Query("SELECT * FROM ORDERS ORDER BY O_CT DESC LIMIT 100")
	defer rows.Close()

	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search orders fail:%v", err)
		return os, err
	}

	for rows.Next() {
		var o Order
		//rows.Columns
		err = rows.Scan(&o.Id, &o.GoodId, &o.GoodCnt, &o.Desc, &o.CusName, &o.CusTel, &o.CusAddr, &o.Status, &o.CreateTime, &o.PayTime, &o.Sum, &o.ShipNo, &o.ShipCom)
		if err != nil {
			glog.V(1).Infof("[DEBUG:] Scan orders fail:%v", err)
			return os, err
		}
		os = append(os, o)
	}
	return os, nil
}

func GetOrdersByStatus(s int64, dbmap *gorp.DbMap) ([]Order, error) {
	var os []Order
	_, err := dbmap.Select(&os, "SELECT * FROM ORDERS WHERE O_ST = ?", s)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search orders by %v fail:%v", s, err)
		return os, err
	}
	return os, nil
}

func GetOrdersByTel(tel int64, dbmap *gorp.DbMap) ([]Order, error) {
	var os []Order
	_, err := dbmap.Select(&os, "SELECT * FROM ORDERS WHERE O_CUS_TEL = ?", tel)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search orders by %v fail:%v", tel, err)
		return os, err
	}
	return os, nil
}

func UpdateOrderFromStatusToStatus(oid string, oldstatus int64, newstatus int64, dbmap *gorp.DbMap) (Order,error) {
	var o Order
	if newstatus > 9 {
		glog.V(1).Infof("[DEBUG:] New status %v invalid", newstatus)
		err := fmt.Errorf("New status %v invalid", newstatus)
		return o,err
	}
	o, err := GetOrderById(oid, dbmap)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by id:%v fail:%v", oid, err)
		return o,err
	}

	if o.Status == newstatus {
		return o,nil
	}

	if o.Status != oldstatus {
		glog.V(1).Infof("[DEBUG:] Current order status is not %v but %v", o.Status, oldstatus)
		err = fmt.Errorf("Current order status is not %v but %v", o.Status, oldstatus)
		return o,err
	}

	o.Status = newstatus
	if newstatus == 5 {
		o.PayTime = NullTime{Time: time.Now(), Valid: true}
	}

	count, err := dbmap.Update(&o)
	if err != nil || count != 1 {
		glog.V(1).Infof("[DEBUG:] Update order %v fail:%v", oid, err)
		return o,err
	}

	return o,nil
}

func UpdateOrderShipping(oid string, shippingNo, shippingCom string, dbmap *gorp.DbMap) (Order,error) {
	var o Order
	o, err := GetOrderById(oid, dbmap)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search order by id:%v fail:%v", oid, err)
		return o,err
	}

	if o.Status != 5 {
		glog.V(1).Infof("[DEBUG:] Current order status is not %v but %v", 5, o.Status)
		err = fmt.Errorf("Current order status is not %v but %v", 5,o.Status)
		return o,err
	}

	o.Status = 3
	o.ShipNo.Scan([]byte(shippingNo))
	o.ShipCom.Scan([]byte(shippingCom))


	count, err := dbmap.Update(&o)
	if err != nil || count != 1 {
		glog.V(1).Infof("[DEBUG:] Update order %v fail:%v", oid, err)
		return o,err
	}

	return o,nil
}
