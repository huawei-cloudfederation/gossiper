package policyengine

import (
	"log"
	"sort"

	"../common"
)

const (
	MIN string = "MIN"
	MAX string = "MAX"
)

type RuleMinMax struct {
	MinOrMax string `json:"MinOrMax"`
}

//this function can be put int the common code
//We dont need to match local only but all dcs
// dcName from policy decision
func GetMatchingConsulDCInfo(dcName string) (*common.DC, bool) {

	//dcDataList cant be null here
	log.Println("GetMatchingConsulDCInfo: called")
	localDCdata, ok := dcDataList[dcName]
	log.Println("GetMatchingConsulDCInfo: Printing the DC maps info", localDCdata, dcName)
	if true != ok {
		log.Println("GetMatchingConsulDCInfo: there is no matching DC in common DC map for", dcName)
		return nil, false
	}
	return localDCdata, true
}

func (this *RuleMinMax) GetnewDCarrangment(policydecision *PolicyDecision) bool {

	if len(policydecision.SortedDCName) == 0 || len(policydecision.SortValue) == 0 {
		log.Println("GetnewDCarrangment: Cannot take a decision and the data set DC list is empty")
		return false
	}

	for index, gossiperDCName := range policydecision.SortedDCName {

		gossiperDcInfo, ok := GetMatchingConsulDCInfo(gossiperDCName)

		if ok != true {
			log.Println("GetnewDCarrangment: failed to get the dc info ")
			continue
		}

		//var cpuPercentage, memPercentage float64

		if gossiperDcInfo.CPU == 0 || gossiperDcInfo.MEM == 0 || gossiperDcInfo.DISK == 0 {
			log.Println("GetnewDCarrangment: Cannot apply policy since the total CPU or total MEM ot DISK in", gossiperDCName, "DC is nil")
			//return false
			continue
		}
		//cpuPercentage = ((gossiperDcInfo.CPU - gossiperDcInfo.Ucpu) / gossiperDcInfo.CPU) * 100
		//memPercentage = ((gossiperDcInfo.MEM - gossiperDcInfo.Umem) / gossiperDcInfo.MEM) * 100

		//TODO: not the right place to decide
		/*if cpuPercentage >= this.RecosurceLimit || memPercentage >= this.RecosurceLimit || diskPercentage >= this.RecosurceLimit {
			log.Println("IsCurrentDCWithInPolicyThershold: DC burst ", cpuPercentage, memPercentage, diskPercentage)
			return false
		}*/
		log.Println("GetnewDCarrangment: The cpupercentage is ", gossiperDcInfo.CPU, " for DC", gossiperDCName)
		policydecision.SortValue[index] = (gossiperDcInfo.CPU)
	}

	sort.Sort(policydecision)
	log.Println("GetnewDCarrangment: Printing the policy dec struct", policydecision)
	return true
}

//sort interface based on the SortValues sort the SortedDCName also
func (this *PolicyDecision) Len() int {
	return len(this.SortValue)
}

func (this *PolicyDecision) Swap(i, j int) {
	this.SortValue[i], this.SortValue[j] = this.SortValue[j], this.SortValue[i]
	this.SortedDCName[i], this.SortedDCName[j] = this.SortedDCName[j], this.SortedDCName[i]

}

//sorts in ascending order
func (this *PolicyDecision) Less(i, j int) bool {

	return this.SortValue[i] < this.SortValue[j]

}

func (this *RuleMinMax) ApplyRule(policydecision *PolicyDecision) bool {

	log.Println("ApplyRule: ApplyRule of the interface to RuleThreshold")

	if ok := this.GetnewDCarrangment(policydecision); ok {

		if this.MinOrMax == "MAX" {
			log.Println("RuleMINMAX: Appying MAX", policydecision.SortValue)
			sort.Sort(sort.Reverse(policydecision))
			log.Println("RuleMINMAX: Appying MAX After reverse sort", policydecision.SortValue)
		} else {
			log.Println("RuleMINMAX: Appying Min", policydecision.SortValue)

		}

		this.SupressoRUnSupress(policydecision)
	}

	//policydecision.SortedDCName[0]

	//

	return true
}

func (this *RuleMinMax) SupressoRUnSupress(policydecision *PolicyDecision) {

	if policydecision.SortedDCName[0] == common.ThisDCName {
		log.Println("SupressoRUnSupress: Current DC will be unspressed", policydecision.SortedDCName[0])
		common.UnSupressFrameWorks()

		//unsupress
	} else {
		//spress
		log.Println("SupressoRUnSupress: Current DC will be supressed", common.ThisDCName, "DC ", policydecision.SortedDCName[0], " Will receive all offers")
		common.SupressFrameWorks()
	}

}
