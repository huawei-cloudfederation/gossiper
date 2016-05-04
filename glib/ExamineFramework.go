package glib

import (
	"log"
	"time"

	"../common"
)

func PerformIntersection() {

	log.Printf("PerformIntersection() Started ")
	FrmWrkLck.Lock()
	CommonFramework = make(map[string]bool)
	for mk, mv := range AllFrameworks {
		for n, _ := range mv {
			isCommon := true
			for sk, sv := range AllFrameworks {
				if mk != sk {
					if _, isvalid := sv[n]; !isvalid {
						isCommon = false
					}
				}
			}
			if isCommon {
				CommonFramework[n] = false
			}
		}
	}

	FrmWrkLck.Unlock()

	common.ToAnon.Lck.Lock()
	for k, _ := range CommonFramework {
		if _, okay := common.ToAnon.M[k]; !okay {
			common.ToAnon.M[k] = true
		}

	}
	//Now remove those frameworks that re not common anymore
	for k, _ := range common.ToAnon.M {
		if _, okay := CommonFramework[k]; !okay {
			delete(common.ToAnon.M, k)
		}
	}
	common.ToAnon.Lck.Unlock()
	log.Printf("PerformIntersection() Finished")
	common.ToAnon.Ch <- true
}

func ExamineFramework() {
	go func() {
		for {
			<-time.After(time.Second * 1)
			log.Printf("Dump AllFramework %v", AllFrameworks)
			log.Printf("Dump CommonFramework %v", CommonFramework)
		}
	}()

	for {
		select {
		case <-time.After(time.Microsecond * 100):
			log.Printf("Performing intersection")
			PerformIntersection()
		}
	}

}
