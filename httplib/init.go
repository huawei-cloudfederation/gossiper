package httplib

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
)

func init() {

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))
	beego.BConfig.CopyRequestBody = true

	beego.Router("/v1/BOOTSTRAP/", &MainController{}, "get:BootStrap")
	beego.Router("/v1/ALLDCSTATUS/", &MainController{}, "get:AllDCStatus")
	beego.Router("/v1/STATUS/", &MainController{}, "get:StatusAll")
	beego.Router("/healthz/", &MainController{}, "get:Healthz")
	beego.Router("/v1/LATENCY/", &MainController{}, "get:LatencyAll")
	beego.Router("/v1/UNSUPRESS/", &MainController{}, "post:UnSupress")
	beego.Router("/v1/THRESHHOLD/", &MainController{}, "post:GetThreshhold")
}
