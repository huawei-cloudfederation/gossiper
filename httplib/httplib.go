package httplib

import (
	"encoding/json"
	"log"

	"github.com/astaxie/beego"

	"../common"
)

type BootStrapResponse struct {
	Name     string
	Country  string
	City     string
	EndPoint string
}

type StatusResponse struct {
	Name              string
	CPU, MEM, DISK    float64 //Total CPU MEM and DISK
	UCPU, UMEM, UDISK float64 //Used CPU MEM and DISK
	OutOfResource     bool
	IsActiveDC        bool
}
type LatencyResponse struct {
	Name string
	Rtt  int64
}

type MainController struct {
	beego.Controller
}

func (this *MainController) LatencyAll() {
	var resp []LatencyResponse
	common.RttOfPeerGossipers.Lck.Lock()
	defer common.RttOfPeerGossipers.Lck.Unlock()
	for k, v := range common.RttOfPeerGossipers.List {
		var c LatencyResponse
		c.Name = k
		c.Rtt = v
		resp = append(resp, c)
	}
	resp_byte, err := json.MarshalIndent(&resp, "", "  ")
	if err != nil {

		log.Printf("Error Marshalling the response")
		this.Ctx.WriteString("Latency Failed")
		return
	}
	this.Ctx.WriteString(string(resp_byte))
}
func (this *MainController) StatusAll() {
	var res StatusResponse

	common.ALLDCs.Lck.Lock()
	defer common.ALLDCs.Lck.Unlock()

	dc, available := common.ALLDCs.List[common.ThisDCName]

	if !available {
		this.Ctx.WriteString("DC information not available")
		log.Printf("DC information not available")
		return
	}

	res.Name = dc.Name
	res.CPU = dc.CPU
	res.MEM = dc.MEM
	res.DISK = dc.DISK
	res.UCPU = dc.Ucpu
	res.UMEM = dc.Umem
	res.UDISK = dc.Udisk
	res.OutOfResource = dc.OutOfResource
	res.IsActiveDC = dc.IsActiveDC

	resp_byte, err := json.MarshalIndent(&res, "", "  ")

	if err != nil {

		log.Printf("Error Marshalling the response")
		this.Ctx.WriteString("Status Failed")
		return
	}

	this.Ctx.WriteString(string(resp_byte))
	log.Printf("HTTP Status %s", string(resp_byte))
}

func (this *MainController) BootStrap() {

	var resp []BootStrapResponse

	for _, v := range common.ALLDCs.List {
		var dc BootStrapResponse
		dc.Name = v.Name
		dc.Country = v.Country
		dc.City = v.City
		dc.EndPoint = v.Endpoint
		resp = append(resp, dc)
	}
	resp_byte, err := json.MarshalIndent(&resp, "", "  ")

	if err != nil {
		log.Println("Something wrong bootstrap api failed %v", err)
		this.Ctx.WriteString("Boot Strap Failed")
		return
	}
	this.Ctx.WriteString(string(resp_byte))
}

func (this *MainController) Healthz() {
	this.Ctx.WriteString("Healthy")
}

func Run(config string) {

	log.Printf("Starting the HTTP server at port %s", config)

	beego.Run(":" + config)

}
