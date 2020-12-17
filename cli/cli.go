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

	"os"
	"strconv"

	"go.dedis.ch/cs438/hw3/client"
	"go.dedis.ch/onet/v3/log"
)

func main() {
	UIPort := flag.String("UIPort", client.DefaultUIPort, "port for  gossip communication with peers")
	msg := flag.String("msg", "i just came to say hello", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message / peer to download the file from")
	share := flag.String("share", "", "list of comma separated files to be shared by the gossiper")
	request := flag.String("request", "", "download the file (MetaFile and all chunks) of the file with this hexadecimal MetaHash")
	filename := flag.String("filename", "", "name used to save the file on local computer")
	keywords := flag.String("keywords", "", "comma separated list of keywords")
	budget := flag.String("budget", "0", "optional search budget; if missing, then the gossiper starts with a budget of 2 and increases it")

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

	// Parsing given file
	if *share != "" && *request == "" {
		fmt.Println("Parsing given files")
		sendMsg(UIAddr, &client.ClientMessage{Share: *share})
		return
	}

	// Request data block from a given file
	if *filename != "" && *request != "" && *share == "" {
		fmt.Println("Requesting data block from a chunk ", *request, " from ", *dest)
		sendMsg(UIAddr, &client.ClientMessage{FileName: *filename, Request: *request, Destination: *dest})
		return
	}

	// Search request for given keywords
	if *keywords != "" {
		if *budget != "" {
			_, err := strconv.ParseUint(*budget, 10, 32)
			if err != nil {
				fmt.Println("cannot convert budget: ", err)
				os.Exit(1)
			}
		}

		fmt.Println("Sending search request for keywords: ", *keywords)
		sendMsg(UIAddr, &client.ClientMessage{Keywords: *keywords, Budget: *budget})
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
