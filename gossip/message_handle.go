package gossip

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"go.dedis.ch/onet/v3/log"
)

// MessageHandler allow to handle HandlingPackets
type MessageHandler struct {
	chanPackets       chan HandlingPacket
	chanReinvokeQueue chan *ReinvokeRumor
	close             bool

	mutexReinvoke sync.Mutex
	reinvokeMap   map[string]*ReinvokeAddr
}

// HandlingPacket Struct to pass for being handle
type HandlingPacket struct {
	addr *net.UDPAddr
	data *GossipPacket
}

// ReinvokeAddr list of rumors to reinvoke for a specific address
type ReinvokeAddr struct {
	mutex  sync.Mutex
	rumors []*ReinvokeRumor
}

// ReinvokeRumor rumor to reinvoke
type ReinvokeRumor struct {
	addr        *net.UDPAddr
	msg         *RumorMessage
	exceptNodes []string
	timer       *time.Timer
}

// NewMessageHandler create a new message handler
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		chanPackets:       make(chan HandlingPacket, 100),
		chanReinvokeQueue: make(chan *ReinvokeRumor, 100),
		close:             false,
		reinvokeMap:       make(map[string]*ReinvokeAddr),
	}
}

// Run handle the received packets
// return a channel which will be closed when all messages have been processed
func (h *MessageHandler) Run(g *Gossiper, packets chan HandlingPacket) <-chan bool {
	packetHandler := func(done chan bool) {
		defer close(done)
		closePacket, closePackets, closeReinvoke := false, false, false
		for {
			select {
			case packet, ok := <-packets:
				if ok {
					h.handlePacket(g, packet)
				} else {
					closePackets = true
					packets = nil
					if closePacket && closePackets && closeReinvoke {
						return
					}
				}
			case packet, ok := <-h.chanPackets:
				if ok {
					h.handlePacket(g, packet)
				} else {
					closePacket = true
					h.chanPackets = nil
					if closePacket && closePackets && closeReinvoke {
						return
					}
				}
			case reinvoke, ok := <-h.chanReinvokeQueue:
				if ok {
					reinvoke.msg.PropagateRumor(g, reinvoke.addr, reinvoke.exceptNodes)
				} else {
					closeReinvoke = true
					h.chanReinvokeQueue = nil
					if closePacket && closePackets && closeReinvoke {
						return
					}
				}
			}
		}
	}

	done := make(chan bool, 1)
	go packetHandler(done)

	return done
}

// Stop gracefully the runing process
func (h *MessageHandler) Stop() {
	h.close = true
	close(h.chanPackets)

	// Stop all reinvoke timers
	func() {
		h.mutexReinvoke.Lock()
		defer h.mutexReinvoke.Unlock()

		for _, address := range h.reinvokeMap {
			//Switch to local lock
			address.mutex.Lock()
			defer address.mutex.Unlock()

			for _, rumor := range address.rumors {
				rumor.timer.Stop()
			}
			address.rumors = nil
		}
		close(h.chanReinvokeQueue)
	}()
}

func (h *MessageHandler) extractMessage(packet GossipPacket) (interface{}, error) {
	// Check wether the message decoded is valid
	if packet.Status != nil && (packet.Private != nil || packet.Rumor != nil) ||
		packet.Private != nil && packet.Rumor != nil {
		return GossipPacket{}, fmt.Errorf("Invalid packet")
	}

	// Return the message
	if packet.Rumor != nil {
		return packet.Rumor, nil
	} else if packet.Private != nil {
		return packet.Private, nil
	} else if packet.Status != nil {
		return packet.Status, nil
	} else {
		return nil, fmt.Errorf("Unsupported packet type")
	}
}

// HandlePacket handle the packet
func (h *MessageHandler) HandlePacket(g *Gossiper, packet HandlingPacket) error {
	if h.close {
		err := errors.New("Handler is closed")
		return err
	}

	h.chanPackets <- packet
	return nil
}

func (h *MessageHandler) handlePacket(g *Gossiper, packet HandlingPacket) {
	msg := packet.data

	// Update address liste
	g.AddAddresses(packet.addr.String())

	message, err := h.extractMessage(*msg)
	if err != nil {
		log.Printf("Error while extracting message: %s", err)
		// Discard the packet
		return
	}

	// switch message.(type) {
	// case *RumorMessage:
	// 	msg := message.(*RumorMessage)
	// 	if msg.Extra != nil {
	// 		if msg.Extra.PaxosAccept != nil {
	// 			log.Printf("%s PAXOS_ACCEPT origin %s from %s ID %d contents %d", g.identifier, msg.Origin, packet.addr.String(), msg.Extra.PaxosAccept.ID, msg.Extra.PaxosAccept.PaxosSeqID)
	// 		} else if msg.Extra.PaxosPrepare != nil {
	// 			log.Printf("%s PAXOS_PREPARE origin %s from %s ID %d contents %d", g.identifier, msg.Origin, packet.addr.String(), msg.ID, msg.Extra.PaxosPrepare.PaxosSeqID)
	// 		} else if msg.Extra.PaxosPromise != nil {
	// 			log.Printf("%s PAXOS_PROMISE origin %s from %s ID %d contents %d", g.identifier, msg.Origin, packet.addr.String(), msg.Extra.PaxosPromise.IDa, msg.Extra.PaxosPromise.PaxosSeqID)
	// 		} else if msg.Extra.PaxosPropose != nil {
	// 			log.Printf("%s PAXOS_PROPOSE origin %s from %s ID %d contents %d", g.identifier, msg.Origin, packet.addr.String(), msg.Extra.PaxosPropose.ID, msg.Extra.PaxosPropose.PaxosSeqID)
	// 		}else if msg.Extra.SwarmInit != nil {
	// 	log.Printf("%s SWARM INIT origin %s from %s PatternID %s Positions %v Targets %v", g.identifier, msg.Origin, packet.addr.String(), msg.Extra.SwarmInit.PatternID, msg.Extra.SwarmInit.DronePos, msg.Extra.SwarmInit.TargetPos)
	// }
	// 	}
	// 	// log.Printf("RUMOR origin %s from %s ID %d contents %s", msg.Origin, packet.addr.String(), msg.ID, msg.Text)

	// case *StatusPacket:
	// 	msg := message.(*StatusPacket)
	// 	var logger strings.Builder
	// 	logger.WriteString("STATUS from " + packet.addr.String())
	// 	for _, s := range msg.Want {
	// 		fmt.Fprintf(&logger, " peer %s nextID %d", s.Identifier, s.NextID)
	// 	}
	// 	log.Printf(logger.String())

	// case *PrivateMessage:
	// 	msg := message.(*PrivateMessage)
	// 	log.Printf("CLIENT MESSAGE %s dest %s", msg.Text, msg.Destination)

	// default:
	// 	log.Printf("UNKNOWN MESSAGE TYPE!\n")
	// 	return
	// }

	// log.Printf("%s PEERS %s", g.identifier, strings.Join(g.GetNodes(), ","))

	// Handle message
	err = g.ExecuteHandler(message, packet.addr)
	if err != nil {
		log.Printf("Error in handler %s", err)
	}
}
