// ========== CS-438 HW2 Skeleton ===========
// *** Implement here the handler for simple message processing ***

package gossip

import (
	"math/rand"
	"net"
	"time"

	"go.dedis.ch/onet/v3/log"
)

// TimeoutMongering time we wait for an ack before reinvoking the rumor
const TimeoutMongering = 10

const ackMessageToSend = 1
const ackMessageToReceive = 2
const ackSynchronised = 2

// Exec is the function that the gossiper uses to execute the handler for a RumorMessage
func (msg *RumorMessage) Exec(g *Gossiper, addr *net.UDPAddr) error {

	// If we receive our own rumor
	if msg.Origin == g.identifier && addr != g.server.Address {
		g.SendMessageTo(*g.CreateStatusMessage(), addr.String())
		return nil
	}

	// Update route
	isRouteRumor := msg.Text == "" && msg.Extra == nil
	g.updateRoute(msg.Origin, addr.String(), msg.ID, !isRouteRumor)

	// Keep track of this rumor
	isNewRumor := false

	nextID, oldNextID := g.trackRumor(msg)
	isNewRumor = msg.ID >= oldNextID

	// Callback + gInWatcher in case of new message
	if nextID > oldNextID {
		track, _ := g.messages.Load(msg.Origin)
		tracking := track.(*messageTracking)
		for i := msg.ID; i < nextID; i++ {
			message, _ := tracking.messages.Load(i)
			rumor := message.(*RumorMessage)
			isPaxosRumor := rumor.Extra != nil

			// Call the callback - Deliver rumor
			if g.callback != nil && g.server.Address != addr && !isRouteRumor {
				g.callback(msg.Origin, GossipPacket{Rumor: rumor}.Copy())
			}
			if isPaxosRumor {
				g.naming.handleExtraMessage(g, rumor.Extra)
			}
		}
	}

	// Send a ack Status that we have receive a message
	if g.server.Address != addr {
		g.SendMessageTo(*g.CreateStatusMessage(), addr.String())
	}

	// If new message
	if isNewRumor {
		msg.PropagateRumor(g, addr, []string{addr.String()})
	}

	return nil
}

// PropagateRumor to a random host
func (msg *RumorMessage) PropagateRumor(g *Gossiper, addr *net.UDPAddr, exceptNodes []string) error {
	// Pick random receiver
	node, err := g.RandomAddress(exceptNodes...)
	if err != nil {
		// No host to send message
		return nil
	}

	// Send the message to this client
	// if len(exceptNodes) > 1 {
	// 	log.Printf("FLIPPED COIN sending rumor to %s", node)
	// }

	exceptNodes = append(exceptNodes, node)

	// Prepare reinvoke
	g.handler.mutexReinvoke.Lock()
	defer g.handler.mutexReinvoke.Unlock()

	if !g.handler.close {
		reinvoke, ok := g.handler.reinvokeMap[addr.String()]
		if !ok {
			reinvoke = &ReinvokeAddr{rumors: make([]*ReinvokeRumor, 0)}
			g.handler.reinvokeMap[addr.String()] = reinvoke
		}
		reinvoke.mutex.Lock()
		defer reinvoke.mutex.Unlock()

		reinvokeRumor := &ReinvokeRumor{
			addr:        addr,
			msg:         msg,
			exceptNodes: exceptNodes,
		}
		reinvoke.rumors = append(reinvoke.rumors, reinvokeRumor)

		reinvokeRumor.timer = time.AfterFunc(TimeoutMongering*time.Second, func() {
			func() {
				reinvoke.mutex.Lock()
				defer reinvoke.mutex.Unlock()
				if g.handler.close {
					// Cancel wake up
					return
				}

				// Remove from list
				i := 0
				for i = 0; i < len(reinvoke.rumors); i++ {
					if reinvoke.rumors[i] == reinvokeRumor {
						break
					}
				}
				reinvoke.rumors[len(reinvoke.rumors)-1], reinvoke.rumors[i] = reinvoke.rumors[i], reinvoke.rumors[len(reinvoke.rumors)-1]
				reinvoke.rumors = reinvoke.rumors[:len(reinvoke.rumors)-1]
			}()

			defer func() {
				recover()
			}()
			g.handler.chanReinvokeQueue <- reinvokeRumor
		})
	}

	// Prepare ack channel
	g.SendMessageTo(GossipPacket{
		Rumor: msg,
	}, node)

	return nil
}

// Exec is the function that the gossiper uses to execute the handler for a StatusMessage
func (msg *StatusPacket) Exec(g *Gossiper, addr *net.UDPAddr) error {
	// Compare vector clock
	messageToSend := make([]PeerStatus, 0)
	messageToReceive := false

	g.messages.Range(func(identifier, track interface{}) bool {
		tracking := track.(*messageTracking)
		id := identifier.(string)
		for _, peer := range msg.Want {
			if peer.Identifier == id {
				if peer.NextID < tracking.nextID {
					messageToSend = append(messageToSend, peer)
				}
				return true
			}
		}
		// identifier not found in msg.Want
		messageToSend = append(messageToSend, PeerStatus{
			Identifier: id,
			NextID:     1,
		})
		return true
	})

	if len(messageToSend) == 0 {
		for _, msg := range msg.Want {
			track, ok := g.messages.Load(msg.Identifier)
			if !ok || track.(*messageTracking).nextID < msg.NextID {
				messageToReceive = true
				break
			}
		}
	}

	ackStatus := ackSynchronised
	if len(messageToSend) > 0 {
		ackStatus = ackMessageToSend
	} else if messageToReceive {
		ackStatus = ackMessageToReceive
		// } else {
		// 	log.Printf("IN SYNC WITH %s", addr.String())
	}

	// Remove reinvocation of confirmed rumors
	func() {
		g.handler.mutexReinvoke.Lock()
		defer g.handler.mutexReinvoke.Unlock()

		if g.handler.close {
			return
		}

		var ok bool
		var reinvoke *ReinvokeAddr

		reinvoke, ok = g.handler.reinvokeMap[addr.String()]
		g.handler.reinvokeMap[addr.String()] = &ReinvokeAddr{rumors: make([]*ReinvokeRumor, 0)}

		//Switch to local lock
		if ok {
			reinvoke.mutex.Lock()
			defer reinvoke.mutex.Unlock()

			for _, rumor := range reinvoke.rumors {
				rumor.timer.Stop()
				if ackStatus == ackSynchronised {
					// Flip a coin
					coin := rand.Float32() < 0.5

					if coin {
						// Continue rumor mongering
						go func() {
							defer func() {
								err := recover()
								if err != nil {
									log.Printf("recover %s", err)
								}
							}()
							g.handler.chanReinvokeQueue <- rumor
						}()
					}
				}
			}
		}
	}()

	if len(messageToSend) > 0 {
		// Send missing message to the sender
		func() {
			for _, packet := range messageToSend {
				track, ok := g.messages.Load(packet.Identifier)
				tracking := track.(*messageTracking)
				var rangeID uint32 = 1
				if ok {
					rangeID = tracking.nextID
				}
				for i := packet.NextID; i < rangeID; i++ {
					message, _ := tracking.messages.Load(i)
					sendingPacket := GossipPacket{
						Rumor: message.(*RumorMessage),
					}
					g.SendMessageTo(sendingPacket, addr.String())
				}
			}
		}()
	} else if messageToReceive {
		// Send Status message to the sender
		g.SendMessageTo(*g.CreateStatusMessage(), addr.String())
	}
	return nil
}

// Exec is the function that the gossiper uses to execute the handler for a PrivateMessage
func (msg *PrivateMessage) Exec(g *Gossiper, addr *net.UDPAddr) error {
	// Update route
	g.updateRoute(msg.Origin, addr.String(), msg.ID, true)

	if g.identifier == msg.Destination {
		// Call the callback
		if g.callback != nil && g.server.Address != addr {
			g.callback(msg.Origin, GossipPacket{Private: msg})
		}

		// We reach the destination
		// log.Printf("PRIVATE origin %s hop-limit %d contents %s", msg.Origin, msg.HopLimit, msg.Text)
	} else if msg.HopLimit > 0 {
		// Forward using our routing table
		destAddr, ok := g.routes.Load(msg.Destination)
		if !ok {
			// Unknown destination, message discarded
			return nil
		}

		msg.HopLimit--
		g.SendMessageTo(GossipPacket{Private: msg}, destAddr.(*RouteStruct).NextHop)
	}
	return nil
}
