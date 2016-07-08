package glib

import (
	"encoding/json"
	"log"
	"time"
	"fmt"

	"../common"
)

type delegate struct {
	glib               *Glib
	GetBroadcastCalled int
}

type LastOORLocal struct {
	Name string //Name of the Datacenter the oor was recived
	TS   int64  //Time stamp we recived last oor

}

var LastOOR LastOORLocal

func (d *delegate) NodeMeta(limit int) []byte {
	log.Printf("Delegate NodeMeta() is called")
	return []byte{}
}

func (d *delegate) NotifyMsg(buf []byte) {
	//Gossipers will recive the message others and update the global map
	log.Printf("Delegate NotifyMsg() is called %s", string(buf))
	var msg Msg
	err := json.Unmarshal(buf, &msg)
	if err != nil {
		log.Printf("Delegate NotifyMsg() unmarshall error %v", err)
		return
	}

	switch msg.Type {
	case "FrameWorkMsG":
		var FW FrameWorkMsG
		msg.Body = &FW
		err := json.Unmarshal(buf, &msg)
		if err != nil {
			log.Printf("Delegate NotifyMsg() unmarshall FrameWorkMsG error %v", err)
			return

		}

		log.Printf("A DC FrameWork Msg %v", msg)

		//First check if the Daacenter entry is available otherwise remove it
		/*
			this_frmwrk, isvalid := AllFrameworks[msg.Name]
			if !isvalid {
				this_frmwrk = make(map[string]bool)
			}
		*/
		this_frmwrk := make(map[string]bool)

		//Loop through the frameworks
		for _, n := range FW.FrameWorks {
			this_frmwrk[n] = false
		}

		FrmWrkLck.Lock()
		AllFrameworks[msg.Name] = this_frmwrk
		FrmWrkLck.Unlock()
		return

	case "OOR":
		var oormsg OutOfResourceMsG
		msg.Body = &oormsg
		err := json.Unmarshal(buf, &msg)
		if err != nil {
			log.Printf("Delegate NotifyMsg() unmarshall OOR error %v", err)
			return
		}
		log.Printf("OORMSG recived %v", msg)
		common.ALLDCs.Lck.Lock()
		defer common.ALLDCs.Lck.Unlock()
		dc, isvalid := common.ALLDCs.List[msg.Name]
		fmt.Println("inside oor common.ThisDCName\n",common.ThisDCName)
		fmt.Println("inside oor msg.Name\n",msg.Name)
		fmt.Println("inside oor isvalid\n",isvalid)
//		msg.Name = "4"
		if isvalid && (msg.Name != common.ThisDCName) {
			fmt.Println("inside isvalid\n")
			if dc.LastOOR != oormsg.TS || LastOOR.Name != msg.Name || LastOOR.TS != oormsg.TS {
				fmt.Println("inside \n")
				dc.OutOfResource = oormsg.OOR
				dc.LastOOR = oormsg.TS
				log.Printf("A DC reported OOR %v", msg)
				go func() {
					time.Sleep(500 * time.Millisecond)
					common.TriggerPolicyCh(true) 
				}()
				LastOOR.Name = msg.Name
				LastOOR.TS = oormsg.TS
			}
		} else {
			log.Printf("Invalid DC name not available in the map %v", msg)
		}
		diff_time := dc.LastUpdate - dc.LastOOR
		if diff_time < 2 {
			// This is a new OOR broadcast we neeed to reboradcast it
			log.Printf("RE boradcasting the OOR message")
			ReGossipOOR(msg)
		}
		return

	case "DC":
		var dc common.DC
		msg.Body = &dc
		err := json.Unmarshal(buf, &msg)
		if err != nil {
			log.Printf("Delegate NotifyMsg() unmarshall Datacenter Mesage error %v", err)
			return
		}

		log.Printf("DC information obtained %v", dc)

		common.ALLDCs.Lck.Lock()
		defer common.ALLDCs.Lck.Unlock()
		common.ALLDCs.List[dc.Name] = &dc
		return
	}

}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	if d.GetBroadcastCalled == 1000 {
		log.Printf("Delegate GetBroadcasts() is called 1000 times")
		d.GetBroadcastCalled = 0
	}
	d.GetBroadcastCalled++
	return d.glib.BC.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	log.Printf("Delegate LocalState() is called")

	return []byte{}
}

func (d *delegate) MergeRemoteState(buf []byte, isJoin bool) {
	log.Printf("Delegate MergeRemoteState() is called")
}
