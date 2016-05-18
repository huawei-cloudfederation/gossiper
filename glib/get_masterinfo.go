package glib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"../common"
)

var (
	AllFrameworks   map[string]map[string]bool //Map of maps for all the frameworks
	CommonFramework map[string]bool            //Common among all the framework
	ThisGlib        *Glib                      //A Common this Glib pointer is required for us to reboradcast OOR message from Delegate
	FrmWrkLck       sync.Mutex                 //Lock for centralized framework
)

func init() {

	AllFrameworks = make(map[string]map[string]bool)
	CommonFramework = make(map[string]bool)
}

func GetListofFrameworks(G *Glib, MasterEP string) {

	resp, err := http.Get(fmt.Sprintf("http://%s/state-summary", MasterEP))

	if err != nil {

		log.Printf("Unable to reach the Master error = %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to read the body error = %v", err)
		return
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		log.Printf("Json Unmarshall error = %v", err)
		return
	}

	frameworks_array, found := data["frameworks"].([]interface{})

	if found != true {
		log.Printf("Unable to find the framework array in the nested json object %v", data)
		return
	}

	//this_frmwrk, exisit := AllFrameworks[G.Name]
	FrmWrkLck.Lock()
	/*
		if !exisit {
			AllFrameworks[G.Name] = make(map[string]bool)
			this_frmwrk = AllFrameworks[G.Name]
		}
	*/
	this_frmwrk := make(map[string]bool)

	for _, frwrk_interface := range frameworks_array {

		frwrk, isvalid := frwrk_interface.(map[string]interface{})
		if !isvalid {
			log.Printf("Malformed json object recived from he master %v", frameworks_array)
			continue
		}
		id, isvalid := frwrk["id"].(string)
		if isvalid != true {
			log.Printf("Unabel to get framework id from the json = %v", frwrk)
			continue
		}
		log.Printf("framework ID %v", id)

		this_frmwrk[id] = true

	}
	AllFrameworks[G.Name] = this_frmwrk
	FrmWrkLck.Unlock()
	G.BroadCast(GossipFrameworks(G.Name))
}

func GetMastersResources(G *Glib, MasterEP string) {

	resp, err := http.Get(fmt.Sprintf("http://%s/metrics/snapshot", MasterEP))

	if err != nil {

		log.Printf("Unable to reach the Master error = %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to read the body error = %v", err)
		return
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		log.Printf("Json Unmarshall error = %v", err)
		return
	}

	var tCPU, uCPU float64
	var tMem, uMem float64
	var tDisk, uDisk float64

	var found bool

	tCPU, found = data["master/cpus_total"].(float64)
	if !found {
		log.Printf("Data not found from master /state/snapshot")
	}
	uCPU, found = data["master/cpus_used"].(float64)
	if !found {
		log.Printf("Data not found from master /state/snapshot")
	}
	tMem, found = data["master/mem_total"].(float64)
	if !found {
		log.Printf("Data not found from master /state/snapshot")
	}
	uMem, found = data["master/mem_used"].(float64)
	if !found {
		log.Printf("Data not found from master /state/snapshot")
	}
	tDisk, found = data["master/disk_total"].(float64)
	if !found {
		log.Printf("Data not found from master /state/snapshot")
	}
	uDisk, found = data["master/disk_used"].(float64)
	if !found {
		log.Printf("Data not found from master /state/snapshot")
	}

	//G.Name
	common.ALLDCs.Lck.Lock()
	var mydc *common.DC
	mydc, available := common.ALLDCs.List[G.Name]

	if !available {
		log.Printf("Our datacenter entry is not found yet")
		mydc = &common.DC{}
		mydc.Name = G.Name
		mydc.Endpoint = common.ThisEP
		mydc.Country = common.ThisCountry
		mydc.City = common.ThisCity
		common.ALLDCs.List[G.Name] = mydc
	}

	log.Println("The values are", tCPU, tMem, tDisk, uCPU, uMem, uDisk)

	mydc.CPU = tCPU
	mydc.MEM = tMem
	mydc.DISK = tDisk
	mydc.Ucpu = uCPU
	mydc.Umem = uMem
	mydc.Udisk = uDisk

	common.ALLDCs.Lck.Unlock()

	CheckThreshold(G, mydc)
	GossipDCInfo(G, mydc)

	return

}

func CheckThreshold(G *Glib, dc *common.DC) {

	var isOOR bool
	isOOR = CheckPercentage(dc.CPU, dc.Ucpu, common.ResourceThresold)
	if !isOOR {
		isOOR = CheckPercentage(dc.MEM, dc.Umem, common.ResourceThresold)
		if !isOOR {
			isOOR = CheckPercentage(dc.DISK, dc.Udisk, common.ResourceThresold)
		}
	}

	dc.LastUpdate = time.Now().Unix()

	// we need to gossip the dc info after the OutOfResource is set
	if dc.OutOfResource != isOOR {
		dc.OutOfResource = isOOR
		dc.LastOOR = dc.LastUpdate
		GossipOOR(G, dc, isOOR)
		common.TriggerPolicyCh <- true
		log.Printf("CheckThreshold() Gossiping OOR now")
	}

	//Now Signal the Policy Engine to

}

func CheckPercentage(MAX, USED float64, Threshold int) bool {

	if USED == 0.0 || MAX == 0.0 || Threshold == 0.0 {
		return false
	}

	if (int(USED/MAX) * 100) > Threshold {

		return true
	}

	return false

}

func GossipOOR(G *Glib, dc *common.DC, isOOR bool) {
	var msg Msg
	oormsg := OutOfResourceMsG{OOR: isOOR, TS: dc.LastOOR}

	msg.Name = common.ThisDCName
	msg.Type = "OOR"
	msg.Body = &oormsg

	msg_bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Unable to broadcast DC information to other gosipers Marshall error")
	} else {

		G.BC.Prune(0) //First Clear the Queue
		G.BroadCast(msg_bytes)
	}

}

func ReGossipOOR(msg Msg) {

	msg_bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Unable to broadcast DC information to other gosipers Marshall error")
	} else {

		ThisGlib.BroadCast(msg_bytes)
	}

}

func GossipDCInfo(G *Glib, dc *common.DC) {

	var msg Msg

	msg.Name = G.Name
	msg.Type = "DC"
	msg.Body = dc
	msg_bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Unable to broadcast DC information to other gosipers Marshall error")
	} else {

		G.BroadCast(msg_bytes)
	}
}

func GossipFrameworks(Name string) []byte {

	var msg Msg
	var FW FrameWorkMsG

	msg.Name = Name
	msg.Type = "FrameWorkMsG"

	gmap, isvalid := AllFrameworks[Name]
	if isvalid {
		for name := range gmap {
			FW.FrameWorks = append(FW.FrameWorks, name)
		}
	}

	msg.Body = &FW
	data, _ := json.Marshal(msg)

	return data
}

func CollectMasterData(G *Glib, MasterEP string) {

	//First get the channels initialized
	//var ResourceFrequency time.Duration
	var FrameworkFrequency, ResourceFrequency time.Duration
	ThisGlib = G
	FrameworkFrequency = 2
	ResourceFrequency = 1010
	getFrameWorkCh := time.After(time.Second * FrameworkFrequency)
	getMasterResourceCh := time.After(time.Millisecond * ResourceFrequency)

	for {
		select {
		case <-getFrameWorkCh:
			getFrameWorkCh = time.After(time.Second * FrameworkFrequency)
			GetListofFrameworks(G, MasterEP)

		case <-getMasterResourceCh:
			getMasterResourceCh = time.After(time.Millisecond * ResourceFrequency)
			GetMastersResources(G, MasterEP)

		}
	}

	log.Fatalf("glib CollectMasterData() finsihed")
}
