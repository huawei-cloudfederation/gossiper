package policyengine

// This interface defines the behaviour of a rule
import (
	"encoding/json"
	"log"
	"sort"
	"time"

	"../common"
)

// All rule typr must implement this rule
type RuleInterface interface {
	ApplyRule(*PolicyDecision) bool
}

// A Group of rules which forms a policy
//
type Policy struct {
	Name  string `json:"Name"` //name of the policy which will be the key in our case
	Rules []Rule `json:"Rules"`
}

//Rule are applied based on the rule priority
//it represents one single rule
// Note: If all the rules have the same priority picks the one which our sort function returns.
type Rule struct {
	Name string `json:"Name"` //this should hold the name of the datatype/struct which will be used to
	//instiantiate the type

	Priority int         `json:"Priority"`        // it should be between 1-10 this 10 being highest
	Scope    string      `json:"Scope,omitempty"` // global scope/local scope
	Content  interface{} `json:"Content"`         // this represt the actual data for the rule, this can be of any type
}

type PolicyDecision struct {
	//after applying all the policy
	SortedDCName     []string  //will have all the sorted dc index 0 is the one which need to be get offers rest suppressed
	LastDecisionTime time.Time // last time this policy decision was taken
	SortValue        []float64
	// can be later use to refresh the same
}

func NewPolicyDecision() *PolicyDecision {
	return &PolicyDecision{}
}

func (this *PE) ProcessNewPolicy(key string, data []byte) (*Policy, error) {
	tempPolicy := &Policy{}
	err := json.Unmarshal(data, &tempPolicy)

	if err != nil {
		log.Println("processNewPolicy: Unable to unmarshal the processNewPolicy", err)
		return nil, err
	}

	for index, values := range tempPolicy.Rules {
		switch values.Name {
		case "MinMax":
			tempPolicy.Rules[index].Content = &RuleMinMax{}
		case "Threshold":
			tempPolicy.Rules[index].Content = &RuleThreshold{}

		}
	}
	err = json.Unmarshal(data, &tempPolicy)
	if err != nil {
		log.Println("processNewPolicy: Unable to get the content from the json")
		return nil, err
	}

	for index := range tempPolicy.Rules {
		if tempPolicy.Rules[index].Name == "Threshold" {
			log.Println("ProcessNewPolicy: Got a new RuleThreshold")
			ruleThershold, ok := tempPolicy.Rules[index].Content.(RuleInterface)
			if ok {
				ruleThershold.ApplyRule(nil)
				//tempPolicy.Rules[index].Content.ApplyRule(nil)
				log.Println("ProcessNewPolicy: The new Threshold is ", common.ResourceThresold, ruleThershold)
			} else {
				log.Println("ProcessNewPolicy: Unable to process a new RuleThreshold")
			}
		}
	}

	return tempPolicy, nil
}

func (this *Policy) TakeDecision() bool {

	//create a new policy decision

	log.Println("TakeDecision: called")
	var ok bool
	var ruleInterface RuleInterface

	//we need to lock the gloabl common map before we take a decision on the policy
	common.ALLDCs.Lck.Lock()
	defer common.ALLDCs.Lck.Unlock()
	newDecision := NewPolicyDecision()

	newDecision.SortedDCName, newDecision.SortValue, ok = GetValidDCsInfo()
	if ok != true {
		log.Println("TakeDecision: Unable to take a decision sort on DC's failed")
	}
	//apply priority
	/*ok = this.ApplyPriority()
	if ok != true {
		log.Println("TakeDecision: Apply priority failes on Policy")
	}*/
	for _, rule := range this.Rules {
		ruleInterface, ok = rule.Content.(RuleInterface)
		if ok != true {
			log.Println("TakeDecision: unable to get the Rule failed to convert to RuleInterface")
			return false
		}
		//this.ApplyPolicyDecisionOnRules(rule.Content)
		log.Println("TakeDecision: Applying Rules ")
		ok := ruleInterface.ApplyRule(newDecision)
		if ok != true {
			log.Println("TakeDecision: Applying rule failes", rule.Name, newDecision.SortedDCName)
		}

	}
	//newDecision.SortedDCName[0]//will be the curent one to get offers

	return true
}

//GetDCDataInSortedOrderByDcName
// this will sort the dc names in
// NewPolicyDecision
func GetValidDCsInfo() ([]string, []float64, bool) {

	if len(dcDataList) == 0 {
		log.Println("GetValidDCsInfo: the common DC maps is empty")
		//common.SupressFrameWorks()
		log.Fatalf("Mesos master is porbably not started please start gossiper after Mesos Master is started")
		return nil, nil, false
	}

	dcDataSortedList := []string{}
	for key, val := range dcDataList {
		if val.OutOfResource == false {
			dcDataSortedList = append(dcDataSortedList, key)
		} else {
			log.Println("GetValidDCsInfo: excluding the DC", key)
		}
	}
	sort.Strings(dcDataSortedList)
	log.Println("GetValidDCsInfo: Valid DC's  from which the Policy decision is made", dcDataSortedList)
	dcDatasortedValues := make([]float64, len(dcDataSortedList))
	return dcDataSortedList, dcDatasortedValues, true

}
