package httplib

import (
	"encoding/json"
	"log"

	"github.com/astaxie/beego"

	"../common"
//	"strconv"
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

type PErequest struct{
        UnSupress bool `json:"UnSupress"`
}
type SetThreshhold struct{
        Threshhold int 
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

func (this *MainController) AllDCStatus() {
	var res []common.DC

	for _, v := range common.ALLDCs.List {
	var dc common.DC 

	dc.Name = v.Name
	dc.City = v.City
	dc.Country = v.Country
	dc.Endpoint = v.Endpoint
	dc.CPU = v.CPU
	dc.MEM = v.MEM
	dc.DISK = v.DISK
	dc.Ucpu = v.Ucpu
	dc.Umem = v.Umem
	dc.Udisk = v.Udisk
	dc.OutOfResource = v.OutOfResource
	dc.IsActiveDC = v.IsActiveDC
	dc.LastUpdate = v.LastUpdate
	dc.LastOOR = v.LastOOR
	res = append(res, dc)
	}

	resp_byte, err := json.MarshalIndent(&res, "", "  ")

	if err != nil {

		log.Printf("Error Marshalling the response")
		this.Ctx.WriteString("Status Failed")
		return
	}

	this.Ctx.WriteString(string(resp_byte))
	log.Printf("HTTP Status %s", string(resp_byte))
}

func (this *MainController) UnSupress(){
        var data  PErequest
        log.Println("From SupresF called")
        this.Data["UnSupress"] = this.Ctx.Input.Param(":UnSupress")

        err := json.Unmarshal(this.Ctx.Input.RequestBody,&data)
        log.Println(string(this.Ctx.Input.RequestBody),"::",data)
        if err != nil {
		this.Ctx.Output.Body(this.Ctx.Input.RequestBody)
                log.Println(string(this.Ctx.Input.RequestBody),"::",data)
                log.Println("Cannot Unmarshal\n",err)
                 return
        }
		this.Ctx.Output.Body(this.Ctx.Input.RequestBody)
                log.Println(string(this.Ctx.Input.RequestBody),"::",data)
	
        if data.UnSupress {
                log.Println(string(this.Ctx.Input.RequestBody),"::",data)
                common.UnSupressFrameWorks()
        }else {
                common.SupressFrameWorks()
        }
}

func (this *MainController) GetThreshhold(){
        var data  SetThreshhold
        log.Println("From Threshhold called")
        this.Data["Threshhold"] = this.Ctx.Input.Param(":Threshhold")

        err := json.Unmarshal(this.Ctx.Input.RequestBody,&data)
        log.Println(string(this.Ctx.Input.RequestBody),"::",data)
        if err != nil {
                this.Ctx.Output.Body(this.Ctx.Input.RequestBody)
                log.Println(string(this.Ctx.Input.RequestBody),"::",data)
                log.Println("Cannot Unmarshal\n",err)
                 return
        }
                this.Ctx.Output.Body(this.Ctx.Input.RequestBody)
                log.Println(string(this.Ctx.Input.RequestBody),"::",data)

	//common.ResourceThresold,_ = strconv.Atoi(data.Threshhold)
	common.ResourceThresold = data.Threshhold

}

func (this *MainController) Healthz() {
	this.Ctx.WriteString("Healthy")
}

func Run(config string) {

	log.Printf("Starting the HTTP server at port %s", config)

	beego.Run(":" + config)

}
