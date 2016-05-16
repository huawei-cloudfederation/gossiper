package anonlib

import (
	"encoding/binary"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"../common"
)

type anonConnection struct {
	sync.Mutex
}

const (
	TYPE_ACK byte = iota
	TYPE_DATA
	TYPE_BEAT
)

//Some Global variables
var (
	ServerAddr string
	ServerPort string
)

// GetTheClientData builds the frame-id:bool value byte string
func BuildClientFrameData(framid string, enable bool) []byte {

	enableStr := "0"
	if enable == true {
		enableStr = "1"
	}
	tempStr := framid + ":" + enableStr
	return []byte(tempStr)

}

// BuildTheClientDataFromMap locks and read the map and build calls the BuildClientFrameData
func BuildTheClientDataFromMap(M map[string]bool) []byte {

	flag := false

	common.ToAnon.Lck.Lock()
	defer common.ToAnon.Lck.Unlock()

	result := []byte{}
	for key, value := range M {
		if flag == true {
			result = append(result, []byte(" ")...)
		}
		result = append(result, BuildClientFrameData(key, value)...)
		flag = true
	}

	return result
}

func GenerateDataMsg(msgType byte, data []byte, totalMsgCount int) []byte {

	//calculate the total byte lenght
	totalLength := int(len(data) + 4 + 4 + 1)
	//create a dummy byte array
	resultByteArray := make([]byte, totalLength)

	//encode msg type in the first byte
	resultByteArray[0] = msgType
	//encode the total msg length
	binary.BigEndian.PutUint32(resultByteArray[1:5], uint32(len(data)+(totalMsgCount*2)))
	//encode the total frames id sent
	binary.BigEndian.PutUint32(resultByteArray[5:9], uint32(totalMsgCount))
	//next start from 9th index
	copy(resultByteArray[9:], data)
	return resultByteArray
}

func NewaAnonConnection() *anonConnection {
	return &anonConnection{}
}

func ConnectToServer() {

	for {
		log.Println("Starting the client")
		DailtoServer()
		<-time.After(5 * time.Second)
	}
}

func DailtoServer() {

	flagChan := make(chan bool)
	log.Println("DailtoServer: called")
	//Bind to all the network inteface int the system on the same port
	tcpServer, err := net.Dial("tcp", ServerAddr+":"+ServerPort)
	if err != nil {
		log.Println("DailtoServer: Unable to connect to the server", err)
		return
	}
	newAnonClientConn := NewaAnonConnection()
	log.Println("DailtoServer: Starting all the go routine")
	go newAnonClientConn.SendMsg(tcpServer, flagChan)
	go newAnonClientConn.RecvMsg(tcpServer, flagChan)
	go newAnonClientConn.SendHeartBeat(tcpServer, flagChan)
	<-flagChan
	log.Println("DailtoServer: flagChan received should abort here")

	//free the socket

	return

}

// SendMsg will wait on the Common.ToAnon
func (annonClient *anonConnection) SendMsg(conn net.Conn, flagChan chan bool) {
	for {
		<-common.ToAnon.Ch
		result := BuildTheClientDataFromMap(common.ToAnon.M)
		result = GenerateDataMsg(TYPE_DATA, result, len(common.ToAnon.M))
		annonClient.Lock()
		log.Println("SendMsg: trying to send data to client", string(result))
		n, err := conn.Write(result)
		if err != nil {
			log.Println("SednMsg: Unable to send data to client", err)
			annonClient.Unlock()
			flagChan <- true
			return
		}
		log.Printf("SendMsg: Total bytes sent %d bytes", n)
		annonClient.Unlock()
	}

}

func (annonClient *anonConnection) RecvMsg(conn net.Conn, flagChan chan bool) {

	for {
		localClientBuf := make([]byte, 4096)
		_, err := conn.Read(localClientBuf)
		if err != nil {
			log.Println("RecvMsg: Unable to read from the client", err, conn.RemoteAddr().String())
			flagChan <- true
			return
		}
		msgType := localClientBuf[0:1]
		if msgType[0] == TYPE_ACK {
			log.Println("RecvMsg: received the ack from the anon server")
		} else {
			log.Println("RecvMsg: Unknown msg from client", msgType, conn.RemoteAddr().String())
		}
	}

}

func (annonClient *anonConnection) SendHeartBeat(conn net.Conn, flagChan chan bool) {

	for {
		<-time.After(10 * time.Second)
		annonClient.Lock()
		_, err := conn.Write([]byte{TYPE_BEAT})
		annonClient.Unlock()
		if err != nil {
			log.Println("SendHeartBeat: Failed to send Heart Beat msg", err, conn.RemoteAddr().String())
			flagChan <- true
			return
		}
	}
}

func Run(inBindAddress string, inBindPort string) {

	if inBindAddress != "" {
           ServerAddr=strings.Split(inBindAddress,":")[0]
		//ServerAddr = strings.TrimRight(strings.TrimSpace(inBindAddress), ":5050")
	}
	if inBindPort != "" {
		ServerPort = inBindPort
	}
	ServerPort = "5555"

	log.Println("Starting the anonlib in new server ",ServerAddr)
	log.Println("The master anon tcp server is", ServerAddr, ServerPort)

	go ConnectToServer()
}
