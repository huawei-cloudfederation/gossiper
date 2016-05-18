package glib

//Define some message structure for Gossip Payload
//All the messages are going to be json messages
//We are going to have a common message structure and under it we coudl have different sub messages

type Msg struct {
	Name string      //Name of the datacenter from which you are getting this message
	Type string      //Type of the Mesage that is being gossiped FrameWork/OOR/DC
	Body interface{} //A Genral body that can be of any type
}

type FrameWorkMsG struct {
	FrameWorks []string
}

type OutOfResourceMsG struct {
	OOR bool  //Is Datacenter Out of Resource
	TS  int64 //Time stamp of this OOR message
}

//NOt used currently
type DCMsG struct {
	Name     string
	EndPoint string
}
