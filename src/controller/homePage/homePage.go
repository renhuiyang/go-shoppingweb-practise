package homePage

import (
	"github.com/martini-contrib/render"
	//"database/sql"
	//"fmt"
	)

func ShowHomePage(r render.Render){
//	type homePageInfo struct{
//		//top
//		top_goods_pic1 string
//		top_goods_pic2 string
//		top_goods_pic3 string
//		
//		//
//		}
//	r.HTML(200,"homepage",pic)
    r.HTML(200,"Laborelief",nil)
	}

func ShowBuy(r render.Render){
    r.HTML(200,"buy",nil)
	}

func ShowAsk(r render.Render){
    r.HTML(200,"ask",nil)
	}

func ShowAvoiding(r render.Render){
	r.HTML(200,"avoiding",nil)
	}

func ShowHow(r render.Render){
	r.HTML(200,"how",nil)
	}

func ShowFaqs(r render.Render){
	r.HTML(200,"faqs",nil)
	}