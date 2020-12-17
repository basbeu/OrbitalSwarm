package client

import "go.dedis.ch/cs438/hw3/paxos/types"

const DefaultUIPort = "8080" // Port number for exchanging messages with the user interface

type ClientMessage struct {
	Contents    string `json:"contents"`
	Destination string `json:"destination"`
}

// ChainResp is used to transmit GetBlocks() responses
type ChainResp struct {
	Last  string
	Chain map[string]types.Block
}
