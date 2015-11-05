package goods

import (
	"github.com/coopernurse/gorp"
	"github.com/golang/glog"
	. "utils/types"
)

func GetGoodsById(gid string, dbmap *gorp.DbMap) (GoodInfo, error) {
	var ginfo GoodInfo
	err := dbmap.SelectOne(&ginfo, "SELECT * FROM GOODS WHERE G_ID=?", gid)
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search goods by id:%v fail:%v", gid, err)
		return ginfo, err
	}
	return ginfo, nil
}

func GetGoods(dbmap *gorp.DbMap) ([]GoodInfo, error) {
	var ginfo []GoodInfo
	_,err := dbmap.Select(&ginfo, "SELECT * FROM GOODS")
	if err != nil {
		glog.V(1).Infof("[DEBUG:] Search goods fail:%v", err)
		return ginfo, err
	}
	return ginfo, nil
}

func GetGoodFeeById(gid string, dbmap *gorp.DbMap) (float32, error) {
	ginfo, err := GetGoodsById(gid, dbmap)
	if err != nil {
		return 0, err
	}
	return ginfo.Fee, nil
}