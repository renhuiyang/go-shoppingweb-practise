package users

import (
	//"database/sql"
	//"encoding/json"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"net/http"
	//"strings"
	//"strconv"
	. "utils/types"
)

func PostUser(r render.Render, dbmap *gorp.DbMap, res http.ResponseWriter, u User, e binding.Errors) {
	if e != nil {
		r.JSON(http.StatusBadRequest, map[string]string{"message": e[0].Message})
		return
	}
	//check user exists yet or not?
	var count int64
	count,err := dbmap.SelectInt("SELECT count(*) FROM CUSTOMER WHERE CUS_TEL=?",u.PhoneNumber)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search customer by tel:%v fail:%v", u.PhoneNumber, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		return
	}

	if count > 0 {
		glog.V(1).Infof("Customer with tel:%v exists yet", u.PhoneNumber)
		r.JSON(http.StatusConflict, map[string]string{"message": "User with same Tel exists"})
		return
	}

	//insert new user info to db;
	err = dbmap.Insert(&u)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Insert customer %v fail:%v", u, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		return
	}
	r.JSON(200, map[string]string{"message": "SUCCESS"})
}

func GetUser(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	tel := params["tel"]
	//get userinfo
	var u User
	err := dbmap.SelectOne(&u, "SELECT * FROM CUSTOMER WHERE CUS_TEL=?", tel)
	//err := db.QueryRow("SELECT CUS_NAME,CUS_ADDR,CUS_EMAIL,CUS_TEL,CUS_PW FROM CUSTOMER WHERE CUS_TEL=?",tel).Scan(&u.Name,&u.Address,&u.Email,&u.PhoneNumber,&u.Password)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search customer by tel:%v fail:%v", tel, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		return
	}
	r.JSON(200, u)
}
