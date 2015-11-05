package goods

import (
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"net/http"
	"time"
	. "utils/types"
	"utils/goods"
)

func PostGoods(r render.Render, dbmap *gorp.DbMap, res http.ResponseWriter, g GoodInfo, e binding.Errors) {
	if e != nil {
		r.JSON(http.StatusBadRequest, map[string]string{"message": e[0].Message})
		return
	}

	//check user exists yet or not?
	var count int64
	//count,err := dbmap.SelectInt("SELECT count(*) FROM ORDERS WHERE O_ID=?",o.Id)
	row, err := dbmap.Db.Query("SELECT count(*) FROM GOODS WHERE G_ID=?", g.Id)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search goods by id:%v fail:%v", g.Id, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
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
		glog.V(1).Infof("Goods with tel:%v exists yet", g.Id)
		r.JSON(http.StatusConflict, map[string]string{"message": "Order with same Id exists"})
		return
	}

	g.CreateTime = NullTime{Time: time.Now(), Valid: true}
	g.UpdateTime = NullTime{Time: time.Now(), Valid: true}

	//insert new user info to db;
	err = dbmap.Insert(&g)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Insert Goods %v fail:%v", g, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		return
	}

	r.JSON(200, map[string]string{"message": "SUCCESS"})
}

func GetGoodsById(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter) {
	id := params["id"]
	//get userinfo
	ginfo, err := goods.GetGoodsById(id, dbmap)
	if err != nil {
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
	}

	r.JSON(200, ginfo)
}

func GetGoods(r render.Render, dbmap *gorp.DbMap, res http.ResponseWriter) {
	//get userinfo
	ginfos, err := goods.GetGoods(dbmap)
	if err != nil {
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
	}
	
	glog.V(2).Infof("[DEBUG]:", ginfos)

	r.JSON(200, ginfos)
}

func PutGoods(r render.Render, dbmap *gorp.DbMap, params martini.Params, res http.ResponseWriter, g GoodInfo, e binding.Errors) {
	id := params["id"]
	//get userinfo
	_, err := goods.GetGoodsById(id, dbmap)
	if err != nil {
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
	}

	//insert new user info to db;
	_, err = dbmap.Update(&g)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Update Goods %v fail:%v", g, err)
		r.JSON(http.StatusConflict, map[string]string{"message": "DB ERROR"})
		return
	}

	r.JSON(200, map[string]string{"message": "SUCCESS"})
}

