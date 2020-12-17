// ========== CS-438 Project ===========
// *** Implement here the gossiper ***
package gossip

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"strings"
	"time"

	"go.dedis.ch/cs438/hw3/gossip/types"
	"go.dedis.ch/cs438/hw3/gossip/watcher"
	"go.dedis.ch/onet/v3/log"

	"sync"
)

// BaseGossipFactory provides a factory to instantiate a Gossiper
//
// - implements gossip.GossipFactory
type BaseGossipFactory struct{}

// New implements gossip.GossipFactory. It creates a new gossiper.
func (f BaseGossipFactory) New(address, identifier string, antiEntropy int,
	routeTimer int, numParticipant, nodeIndex int, paxosRetry int) (BaseGossiper, error) {
	return NewGossiper(address, identifier, antiEntropy, routeTimer, numParticipant, nodeIndex, paxosRetry)
}

type messageTracking struct {
	messages sync.Map //map[uint32]string
	nextID   uint32
	mutex    sync.Mutex
}

// Gossiper provides the functionalities to handle a distributed gossip
// protocol.
//
// - implements gossip.BaseGossiper
type Gossiper struct {
	inWatcher  watcher.Watcher
	outWatcher watcher.Watcher

	Handlers map[reflect.Type]interface{}

	server  *UDPServer
	handler *MessageHandler
	sending chan<- UDPPacket

	identifier  string
	address     string
	antiEntropy int
	routeTimer  int
	callback    NewMessageCallback

	nodes      map[string]*net.UDPAddr
	mutexNodes sync.RWMutex

	nextID      uint32
	mutexNextID sync.Mutex

	messages sync.Map
	routes   sync.Map // map[string]*RouteStruct

	chanRouteRumorStop  chan bool
	timerRouteRumor     *time.Ticker
	chanAntiEntropyStop chan bool
	timerAntiEntropy    *time.Ticker

	// Folder to store data about indexed files (comes from -sharedir)
	rootSharedData string
	// Folder to store downloaded files (comes from -downdir)
	rootDownloadedFiles string

	chunks        map[string][]byte
	waitingChunks map[string]chan bool

	// Origin -> MetaHash -> filename
	searchIndexes map[string]map[string]string
	// MetaHash -> NodeID -> ChunksIds
	searchResults map[string]map[string][]uint32
	// Signal some news results
	waitingSearch chan bool

	searchRequests map[string]time.Time

	naming       *Naming
	indexedFiles map[string]string
}

// NewGossiper returns a Gossiper that is able to listen to the given address
// and which has the given identifier. The address must be a valid IPv4 UDP
// address. This method can panic if it is not possible to create a
// listener on that address. To run the gossip protocol, call `Run` on the
// gossiper.
func NewGossiper(address, identifier string, antiEntropy int, routeTimer int, numParticipant int,
	nodeIndex int, paxosRetry int) (BaseGossiper, error) {
	// Configs
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	// Validate IP Address
	server, err := NewUDPServer(address)
	if err != nil {
		return nil, err
	}

	// Default value for anti-entropie
	if antiEntropy <= 0 {
		antiEntropy = 10
	}

	// Create gossiper
	g := &Gossiper{
		inWatcher:  watcher.NewSimpleWatcher(),
		outWatcher: watcher.NewSimpleWatcher(),

		Handlers: make(map[reflect.Type]interface{}),
		handler:  NewMessageHandler(),

		identifier:  identifier,
		address:     server.socket.LocalAddr().String(),
		antiEntropy: antiEntropy,
		routeTimer:  routeTimer,
		callback:    nil,

		server:  server,
		sending: nil,

		nextID:              1,
		chanRouteRumorStop:  make(chan bool, 1),
		chanAntiEntropyStop: make(chan bool, 1),

		nodes: make(map[string]*net.UDPAddr),

		naming: NewNaming(numParticipant, nodeIndex, paxosRetry),
	}

	// Register handler
	err = g.RegisterHandler(&SimpleMessage{})
	if err != nil {
		return nil, err
	}
	err = g.RegisterHandler(&RumorMessage{})
	if err != nil {
		return nil, err
	}
	err = g.RegisterHandler(&StatusPacket{})
	if err != nil {
		return nil, err
	}
	err = g.RegisterHandler(&PrivateMessage{})
	if err != nil {
		return nil, err
	}

	log.Printf("Gossiper create %s at %s", g.identifier, g.address)
	return g, nil
}

func (g *Gossiper) decodePacket(chPacket <-chan UDPPacket) chan HandlingPacket {
	ch := make(chan HandlingPacket, 1024)
	go func() {
		defer close(ch)
		for packet := range chPacket {
			var decodedPacket GossipPacket
			err := json.Unmarshal(packet.data, &decodedPacket)
			if err != nil {
				log.Printf("Discard decoded packet, %s", err)
			} else {
				ch <- HandlingPacket{
					data: &decodedPacket,
					addr: packet.addr,
				}
			}
		}
	}()
	return ch
}

// Run implements gossip.BaseGossiper. It starts the listening of UDP datagrams
// on the given address and starts the antientropy. This is a blocking function.
func (g *Gossiper) Run(ready chan struct{}) {
	//Start server
	listener, sender, handlingFinished := g.server.Run()

	handlerClosed := g.handler.Run(g, g.decodePacket(listener))
	g.sending = sender

	// Ready to receive packets -> close ready channel
	close(ready)

	// Anti-entropy
	if g.antiEntropy > 0 {
		g.timerAntiEntropy = time.NewTicker(time.Second * time.Duration(g.antiEntropy))

		sendStatusMessage := func() {
			// g.BroadcastMessage(*g.CreateStatusMessage())
			if addr, err := g.RandomAddress(""); err == nil {
				g.SendMessageTo(*g.CreateStatusMessage(), addr)
			}
		}
		go func() {
			for {
				select {
				case <-g.chanAntiEntropyStop:
					return
				case <-g.timerAntiEntropy.C:
					sendStatusMessage()
				}
			}
		}()
	}

	// Start route rumor
	if g.routeTimer > 0 {
		g.timerRouteRumor = time.NewTicker(time.Second * time.Duration(g.routeTimer))

		msg := GossipPacket{
			Rumor: &RumorMessage{
				Origin: g.identifier,
				ID:     0,
				Text:   "",
			},
		}

		sendEmptyRouteRumor := func() {
			g.mutexNextID.Lock()
			id := g.nextID
			g.nextID++
			g.mutexNextID.Unlock()

			msg.Rumor.ID = id
			// To one address
			g.handler.HandlePacket(g, HandlingPacket{data: &msg, addr: g.server.Address})
		}

		go func() {
			sendEmptyRouteRumor()
			for {
				select {
				case <-g.chanRouteRumorStop:
					return
				case <-g.timerRouteRumor.C:
					sendEmptyRouteRumor()
				}
			}
		}()
	}

	// Index files in DownloadedFiles
	go func() {
		files, err := ioutil.ReadDir(g.rootDownloadedFiles)
		if err != nil {
			return
		}
		var strFiles []string
		for _, f := range files {
			strFiles = append(strFiles, f.Name())
		}

		if len(strFiles) > 0 {
			g.IndexShares(strings.Join(strFiles, ","))
		}
	}()

	// Connect close handling to handler close event
	handlingFinished <- <-handlerClosed
}

// Stop implements gossip.BaseGossiper. It closes the UDP connection
func (g *Gossiper) Stop() {
	if g.antiEntropy > 0 {
		g.timerAntiEntropy.Stop()
		close(g.chanAntiEntropyStop)
	}
	if g.routeTimer > 0 {
		g.timerRouteRumor.Stop()
		close(g.chanRouteRumorStop)
	}

	g.handler.Stop()
	g.server.Stop()
	// log.Printf("Gossiper closed gracefully")
}

// GetBlocks returns all the blocks added so far. Key should be hexadecimal
// representation of the block's hash. The first return is the hexadecimal
// hash of the last block.
func (g *Gossiper) GetBlocks() (string, map[string]types.Block) {
	return g.naming.GetBlocks()
}

func (g *Gossiper) updateRoute(destination, nextHop string, lastID uint32, officialDsdvUpdate bool) {
	if g.address == nextHop {
		// log.Printf("Try to update our own route %s VS %s", g.address, nextHop)
		return
	}

	route, ok := g.routes.LoadOrStore(destination, &RouteStruct{
		NextHop: nextHop,
		LastID:  lastID,
	})
	if !ok {
		route.(*RouteStruct).NextHop = nextHop
		route.(*RouteStruct).LastID = lastID
	}
	// if officialDsdvUpdate {
	// log.Printf("DSDV %s %s", destination, nextHop)
	// }
}

func (g *Gossiper) trackRumor(msg *RumorMessage) (uint32, uint32) {
	trackMessage := func(tracking *messageTracking) (uint32, uint32) {
		tracking.mutex.Lock()
		defer tracking.mutex.Unlock()
		ID := tracking.nextID

		tracking.messages.Store(msg.ID, msg)
		if tracking.nextID != msg.ID {
			return ID, ID
		}
		tracking.nextID++

		// Find next ID
		for i := tracking.nextID; ; i++ {
			if _, ok := tracking.messages.Load(i); ok {
				tracking.nextID++
				continue
			}
			return tracking.nextID, ID
		}
	}

	ID := uint32(1)
	if msg.ID == 1 {
		ID = 2
	}
	tracking := &messageTracking{
		nextID: ID,
	}
	tracking.messages.Store(msg.ID, msg)

	track, loaded := g.messages.LoadOrStore(msg.Origin, tracking)
	if loaded {
		tracking = track.(*messageTracking)
		tracking.messages.Store(msg.ID, msg)
		return trackMessage(tracking)
	}
	return ID, 1
}

// AddSimpleMessage implements gossip.BaseGossiper. It takes a text that will be
// spread through the gossip network with the identifier of g.
func (g *Gossiper) AddSimpleMessage(text string) {
	// Logging
	// log.Printf("CLIENT MESSAGE %s", text)
	// log.Printf("PEERS %s", strings.Join(g.GetNodes(), ","))

	msg := &GossipPacket{
		Simple: &SimpleMessage{
			OriginPeerName: g.identifier,
			RelayPeerAddr:  g.address,
			Contents:       text,
		},
	}

	// Simply call handler
	g.handler.HandlePacket(g, HandlingPacket{
		data: msg,
		addr: g.server.Address,
	})
}

// AddPrivateMessage sends the message to the next hop.
func (g *Gossiper) AddPrivateMessage(text, dest, origin string, hoplimit int) {
	msg := &GossipPacket{
		Private: &PrivateMessage{
			Origin:      origin,
			ID:          0,
			Text:        text,
			Destination: dest,
			HopLimit:    hoplimit,
		},
	}

	g.handler.HandlePacket(g, HandlingPacket{
		data: msg,
		addr: g.server.Address,
	})
}

// AddMessage takes a text that will be spread through the gossip network
// with the identifier of g. It returns the ID of the message
func (g *Gossiper) AddMessage(text string) uint32 {
	// log.Printf("CLIENT MESSAGE %s", text)

	// Generate next ID
	g.mutexNextID.Lock()
	id := g.nextID
	g.nextID++
	g.mutexNextID.Unlock()

	msg := &GossipPacket{
		Rumor: &RumorMessage{
			Origin: g.identifier,
			ID:     id,
			Text:   text,
		},
	}
	// Simply dispatch message
	g.handler.HandlePacket(g, HandlingPacket{
		data: msg,
		addr: g.server.Address,
	})

	return id
}

// AddExtraMessage allow to send some paxos message
func (g *Gossiper) AddExtraMessage(paxosMsg *types.ExtraMessage) uint32 {
	// Generate next ID
	g.mutexNextID.Lock()
	id := g.nextID
	g.nextID++
	g.mutexNextID.Unlock()

	msg := &GossipPacket{
		Rumor: &RumorMessage{
			Origin: g.identifier,
			ID:     id,
			Extra:  paxosMsg,
		},
	}

	// Simply dispatch message
	g.handler.HandlePacket(g, HandlingPacket{
		data: msg,
		addr: g.server.Address,
	})

	return id
}

// AddAddresses implements gossip.BaseGossiper. It takes any number of node
// addresses that the gossiper can contact in the gossiping network.
func (g *Gossiper) AddAddresses(addresses ...string) error {
	for _, a := range addresses {
		// Check if address is valid
		addr, err := net.ResolveUDPAddr("udp", a)
		if err != nil {
			return err
		}

		// Filter address to check that ours is not inside
		if a != g.address {
			func() {
				g.mutexNodes.Lock()
				defer g.mutexNodes.Unlock()
				g.nodes[a] = addr
			}()
		}
	}

	return nil
}

// RandomAddress Return a random address from the known nodes
func (g *Gossiper) RandomAddress(exceptAddresses ...string) (string, error) {
	g.mutexNodes.RLock()
	defer g.mutexNodes.RUnlock()

	nodes := make([]string, 0, len(g.nodes))

NodeLoop:
	for n := range g.nodes {
		for _, addr := range exceptAddresses {
			if addr == n {
				continue NodeLoop
			}
		}
		nodes = append(nodes, n)
	}

	if len(nodes) <= 0 {
		return "", errors.New("No other known hosts")
	}
	node := nodes[rand.Intn(len(nodes))]
	return node, nil
}

// GetNodes implements gossip.BaseGossiper. It returns the list of nodes this
// gossiper knows currently in the network.
func (g *Gossiper) GetNodes() []string {
	g.mutexNodes.RLock()
	defer g.mutexNodes.RUnlock()

	nodes := make([]string, 0, len(g.nodes))
	for n := range g.nodes {
		nodes = append(nodes, n)
	}

	return nodes
}

// GetDirectNodes implements gossip.BaseGossiper. It returns the list of nodes whose routes are known to this node
func (g *Gossiper) GetDirectNodes() []string {
	routes := make([]string, 0)
	g.routes.Range(func(id, _ interface{}) bool {
		routes = append(routes, id.(string))
		return true
	})
	return routes
}

// SetIdentifier implements gossip.BaseGossiper. It changes the identifier sent
// with messages originating from this gossiper.
func (g *Gossiper) SetIdentifier(id string) {
	g.identifier = id
}

// GetIdentifier implements gossip.BaseGossiper. It returns the currently used
// identifier for outgoing messages from this gossiper.
func (g *Gossiper) GetIdentifier() string {
	return g.identifier
}

// GetRoutingTable implements gossip.BaseGossiper. It returns the known routes.
func (g *Gossiper) GetRoutingTable() map[string]*RouteStruct {
	routes := make(map[string]*RouteStruct)
	g.routes.Range(func(id, route interface{}) bool {
		routes[id.(string)] = route.(*RouteStruct)
		return true
	})
	return routes
}

// GetLocalAddr implements gossip.BaseGossiper. It returns the address
// (ip:port as a string) currently used to send to and receive messages
// from other peers.
func (g *Gossiper) GetLocalAddr() string {
	return g.address
}

// AddRoute updates the gossiper's routing table by adding a next hop for the given
// peer node
func (g *Gossiper) AddRoute(peerName, nextHop string) {
	g.updateRoute(peerName, nextHop, 0, true)
}

// RegisterCallback implements gossip.BaseGossiper. It sets the callback that
// must be called each time a new message arrives.
func (g *Gossiper) RegisterCallback(m NewMessageCallback) {
	// Assuming that the callbacks are thread-safe
	g.callback = m
}

// Watch implements gossip.BaseGossiper. It returns a chan populated with new
// incoming packets
func (g *Gossiper) Watch(ctx context.Context, fromIncoming bool) <-chan CallbackPacket {
	w := g.inWatcher

	if !fromIncoming {
		w = g.outWatcher
	}

	o := &observer{
		ch: make(chan CallbackPacket),
	}

	w.Add(o)

	go func() {
		<-ctx.Done()
		// empty the channel
		o.terminate()
		w.Remove(o)
	}()

	return o.ch
}

// - implements watcher.observable
type observer struct {
	sync.Mutex
	ch      chan CallbackPacket
	buffer  []CallbackPacket
	closed  bool
	running bool
}

func (o *observer) Notify(i interface{}) {
	o.Lock()
	defer o.Unlock()

	if o.closed {
		return
	}

	if o.running {
		o.buffer = append(o.buffer, i.(CallbackPacket))
		return
	}

	select {
	case o.ch <- i.(CallbackPacket):

	default:
		// The buffer size is not controlled as we assume the event will be read
		// shortly by the caller.
		o.buffer = append(o.buffer, i.(CallbackPacket))

		o.checkSize()

		o.running = true

		go o.run()
	}
}

func (o *observer) run() {
	for {
		o.Lock()

		if len(o.buffer) == 0 {
			o.running = false
			o.Unlock()
			return
		}

		msg := o.buffer[0]
		o.buffer = o.buffer[1:]

		o.Unlock()

		// Wait for the channel to be available to writings.
		o.ch <- msg
	}
}

func (o *observer) checkSize() {
	const warnLimit = 1000
	if len(o.buffer) >= warnLimit {
		log.Warn("Observer queue is growing insanely")
	}
}

func (o *observer) terminate() {
	o.Lock()
	defer o.Unlock()

	o.closed = true

	if o.running {
		o.running = false
		o.buffer = nil

		// Drain the message in transit to close the channel properly.
		select {
		case <-o.ch:
		default:
		}
	}

	close(o.ch)
}
