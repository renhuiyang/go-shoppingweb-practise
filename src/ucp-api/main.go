package main

import (
	"fmt"
	"database/sql"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/robfig/cron"
	"log"
	"net/http"
	"os"
	//"html/template"
	"controller/alipay"
	"controller/authService"
	"controller/goods"
	//"controller/homePage"
	"controller/orders"
	"controller/users"
	"utils/authentication"
	. "utils/types"
	//"encoding/json"
)

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func initDb() *gorp.DbMap {
	name := fmt.Sprintf("%v",Config.DB.Dbtype)
	server := fmt.Sprintf("%v:%v@%v/%v?charset=utf8&parseTime=true",Config.DB.User,Config.DB.Password,Config.DB.Server,Config.DB.Dbname)
	//fmt.Println(name+"<>"+server)
	glog.V(1).Infof("[Config] %v",Config)
	db, err := sql.Open(name, server)
	//db, err := sql.Open("mysql", "root:bst321@tcp(localhost:3306)/shopping?charset=utf8&parseTime=true")
	checkErr(err, "sql.Open failed")

	Dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	Dbmap.AddTableWithName(User{}, "CUSTOMER").SetKeys(false, "CUS_TEL")
	Dbmap.AddTableWithName(Order{}, "ORDERS").SetKeys(false, "O_ID")
	Dbmap.AddTableWithName(GoodInfo{}, "GOODS").SetKeys(false, "G_ID")

	err = Dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return Dbmap
}

func readConfig(path string, config *TomlConfig) error {
	if _, err := toml.DecodeFile(path, config); err != nil {
		glog.V(1).Infof("[DEBUG] read configure file fail:%v!",err.Error())
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	flag.Lookup("log_dir").Value.Set("/tmp/log")
	flag.Lookup("v").Value.Set("4")

	err := readConfig("/tmp/service.conf", &Config)
	if err != nil {
		return
	}

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory:  "html/templates",
		Extensions: []string{".tmpl"},
	}))
	m.Use(martini.Static("html",martini.StaticOptions{SkipLogging:false}))

	//store := sessions.NewCookieStore([]byte("ASKJJIJLJKHLSDIOICCCMMNNV"))
	//store.Options(sessions.Options{MaxAge:0})
	//m.Use(sessions.Sessions("my_session", store))
	//m.Use(sessionauth.SessionUser(GenerateAnonymousUser))

	//sessionauth.RedirectUrl = "/login"
	//sessionauth.RedirectParam = "next"
    
	DbMap = initDb()

	defer DbMap.Db.Close()

	DbMap.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))

	m.Group("/user", func(r martini.Router) {
		r.Post("", binding.Json(User{}), users.PostUser)
		r.Get("/:tel", authentication.RequireTokenAuthentication, users.GetUser)
	})

	m.Post("/login", binding.Json(User{}), authService.PostLogin)
	m.Get("/logout", authentication.RequireTokenAuthentication, authService.Logout)

	m.Group("/order", func(r martini.Router) {
		r.Post("/alipay/:type", binding.Json(Order{}), orders.PostOrder)
		r.Get("/check/:tel/:id", orders.GetOrder)
		r.Get("",authentication.RequireTokenAuthentication,orders.GetOrders)
		r.Get("/status/:status",authentication.RequireTokenAuthentication,orders.GetOrdersByStatus)
		r.Get("/tel/:tel",authentication.RequireTokenAuthentication,orders.GetOrdersByTel)
		r.Put("/updatestatus/:id/:oldstatus/:newstatus",authentication.RequireTokenAuthentication,orders.PutOrderStatus)
		r.Put("/shipping/:id/:shippingNo/:shippingCom",authentication.RequireTokenAuthentication,orders.PutOrderShipping)
		r.Get("/search/:id",orders.GetOrderShipping)
		r.Get("/test",orders.Test)
	})

	m.Group("/alipay", func(r martini.Router) {
		r.Post("/notify", alipay.AlipayNotify)
		r.Get("/return", alipay.AlipayReturn)
		r.Get("/testreturn",alipay.Test)
		r.Post("/testnotify",alipay.TestNotify)
	})

	m.Group("/goods", func(r martini.Router) {
		r.Post("", binding.Json(GoodInfo{}), goods.PostGoods)
		r.Get("/", goods.GetGoods)
		r.Get("/:id", goods.GetGoodsById)
		r.Put("/:id", binding.Json(GoodInfo{}), goods.PutGoods)
	})
	
	c := cron.New()
	c.AddFunc("0 0 3 * * *", orders.CheckOrderStatus)
	c.Start()

	m.Map(DbMap)
	//serv := fmt.Sprintf(":%d", config.GetHttpPort())
	http.ListenAndServe(":9092", m)
}
