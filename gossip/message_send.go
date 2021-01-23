package gossip

import (
	"encoding/json"
	"errors"
	"net"

	"go.dedis.ch/onet/v3/log"
)

func (g *Gossiper) encodePacket(packet GossipPacket) ([]byte, error) {
	encodedPacket, err := json.Marshal(packet)
	if err != nil {
		return make([]byte, 0), err
	}

	return encodedPacket, nil
}

// BroadcastMessage broadcast a message to all known hosts
func (g *Gossiper) BroadcastMessage(msg GossipPacket) {
	g.BroadcastMessageExcept(msg, "")
}

// BroadcastMessageExcept broadcast a message to all known hosts except to the given host
func (g *Gossiper) BroadcastMessageExcept(msg GossipPacket, exceptAddresses string) {
	jsonData, err := g.encodePacket(msg)
	if err != nil {
		// Discard invalid message
		return
	}

	for node, addr := range g.nodes {
		if node != exceptAddresses {
			defer recover()
			g.sending <- UDPPacket{data: jsonData, addr: addr}
		}
	}
}

// SendMessageTo send a given message to a given host
func (g *Gossiper) SendMessageTo(msg GossipPacket, addr string) {
	packet, ok := g.encodePacket(msg)
	if ok != nil {
		log.Printf("Discard invalid message while encoding")
		return
	}

	address, err := g.resolveAddresse(addr)
	if err != nil {
		// Error while resolving address
		log.Printf("ERROR While resolving address")
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Recover from sender closed")
		}
	}()
	g.sending <- UDPPacket{data: packet, addr: address}
}

// CreateStatusMessage send a status message to the given address
func (g *Gossiper) CreateStatusMessage() *GossipPacket {
	msg := StatusPacket{
		Want: make([]PeerStatus, 0),
	}
	g.messages.Range(func(identifier, track interface{}) bool {
		msg.Want = append(msg.Want, PeerStatus{
			NextID:     track.(*messageTracking).nextID,
			Identifier: identifier.(string),
		})
		return true
	})

	return &GossipPacket{Status: &msg}
}

func (g *Gossiper) resolveAddresse(addr string) (*net.UDPAddr, error) {
	g.mutexNodes.Lock()
	defer g.mutexNodes.Unlock()
	address, found := g.nodes[addr]

	if !found {
		address, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, errors.New("Unable to resolve address")
		}
		g.nodes[addr] = address
	}
	return address, nil
}
