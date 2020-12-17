// ========== CS-438 HW3 Skeleton ===========
// *** Implement here the CLI client ***

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.dedis.ch/cs438/hw3/client"
	"go.dedis.ch/onet/v3/log"
)

func main() {
	UIPort := flag.String("UIPort", client.DefaultUIPort, "port for  gossip communication with peers")
	msg := flag.String("msg", "i just came to say hello", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message / peer to download the file from")

	flag.Parse()

	UIAddr := "http://127.0.0.1:" + *UIPort
	fmt.Println("client contacts", UIAddr, "with msg", *msg)

	if dest != nil {
		fmt.Println("Destination is:", *dest)
	}

	if *msg != "" && *share == "" && *request == "" {
		fmt.Println("Sending private message or normal")
		sendMsg(UIAddr, &client.ClientMessage{Contents: *msg, Destination: *dest})
		return
	}
}

func encodePacket(msg *client.ClientMessage) []byte {
	encodedMsg, err := json.Marshal(msg)
	if err != nil {
		log.Panic(err)
	}
	return encodedMsg
}

func decodePacket(msg []byte) *client.ClientMessage {
	var message client.ClientMessage
	err := json.Unmarshal(msg, &message)

	if err != nil {
		log.Panic(err)
	}

	return &message
}

// sendMsg json encodes the packet and sends it as an UDP datagram
// to the given address + "/message"
// Note that it must be able to handle ClientMessage.Destination now
func sendMsg(address string, p *client.ClientMessage) {
	var jsonStr = encodePacket(p)
	req, err := http.NewRequest("POST", address+"/message", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if len(body) > 0 {
		decoded := decodePacket(body)
		fmt.Println("Response Body: " + decoded.Contents)
	}
}
