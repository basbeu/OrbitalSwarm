// ========== CS-438 orbitalswarm Skeleton ===========

package drone

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"go.dedis.ch/cs438/orbitalswarm/drone/mapping"
	"go.dedis.ch/cs438/orbitalswarm/paxos"
	"gonum.org/v1/gonum/spatial/r3"

	//"go.dedis.ch/cs438/orbitalswarm/client"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"golang.org/x/xerrors"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type key int

const (
	requestIDKey key = 0
)

// ClientMessage internally represents messages comming from the client
type ClientMessage struct {
	Contents    string `json:"contents"`
	Destination string `json:"destination"`
}

// Drone is responsible to be the glue between the gossiping protocol and
// the ui, dispatching responses and messages etc
type Drone struct {
	sync.Mutex
	droneID       uint32
	uiAddress     string
	identifier    string
	gossipAddress string
	gossiper      *gossip.Gossiper
	cliConn       net.Conn
	messages      []CtrlMessage
	position      r3.Vec
	target        r3.Vec
	mapping       *mapping.Mapping
	targetsMapper mapping.TargetsMapper
	naming        *paxos.Naming // TO TEST PAXOS with naming
}

// CtrlMessage internal representation of messages for the controller of the UI
type CtrlMessage struct {
	Origin string
	ID     uint32
	Text   string
}

// NewDrone returns the controller that sets up the gossiping state machine
// as well as the web routing. It uses the same gossiping address for the
// identifier.
func NewDrone(droneID uint32, identifier, uiAddress, gossipAddress string,
	g *gossip.Gossiper, addresses []string, position r3.Vec, targetsMapper mapping.TargetsMapper, mapping *mapping.Mapping, naming *paxos.Naming) *Drone {

	c := &Drone{
		droneID:       droneID,
		identifier:    identifier,
		uiAddress:     uiAddress,
		gossipAddress: gossipAddress,
		gossiper:      g,
		mapping:       mapping,
		position:      position,
		targetsMapper: targetsMapper,
		naming:        naming, // TO TEST PAXOS with naming
	}
	g.AddAddresses(addresses...)

	g.RegisterCallback(c.HandleGossipMessage)
	return c
}

// Run ...
func (c *Drone) Run() {
	logger := log.With().Timestamp().Str("role", "http proxy").Logger()

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	r := mux.NewRouter()
	r.Methods("GET").Path("/message").HandlerFunc(c.GetMessage)
	r.Methods("POST").Path("/message").HandlerFunc(c.PostMessage)

	r.Methods("GET").Path("/origin").HandlerFunc(c.GetDirectNode)

	r.Methods("GET").Path("/node").HandlerFunc(c.GetNode)
	r.Methods("POST").Path("/node").HandlerFunc(c.PostNode)

	r.Methods("GET").Path("/id").HandlerFunc(c.GetIdentifier)
	r.Methods("POST").Path("/id").HandlerFunc(c.SetIdentifier)

	r.Methods("GET").Path("/routing").HandlerFunc(c.GetRoutingTable)
	r.Methods("POST").Path("/routing").HandlerFunc(c.AddRoute)

	r.Methods("GET").Path("/address").HandlerFunc(c.GetLocalAddr)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./drone/static/")))

	server := &http.Server{
		Addr:    c.uiAddress,
		Handler: tracing(nextRequestID)(logging(logger)(r)),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// GetMessage returns all messages seen so far as json encoded Message
// XXX lot of optimizations to be done here
func (c *Drone) GetMessage(w http.ResponseWriter, r *http.Request) {
	c.Lock()
	defer c.Unlock()
	log.Info().Msgf("These are the msgs %v", c.messages)
	if err := json.NewEncoder(w).Encode(c.messages); err != nil {
		log.Err(err)
		http.Error(w, "could not encode json", http.StatusInternalServerError)
		return
	}
	log.Info().Msg("GUI request for the messages received by the gossiper")
}

// PostMessage with text in the body as raw string
func (c *Drone) PostMessage(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POSTING MESSAGE")
	c.Lock()
	defer c.Unlock()
	text, ok := readString(w, r)
	if !ok {
		log.Err(xerrors.New("failed to read string"))
		return
	}
	message := ClientMessage{}
	err := json.Unmarshal([]byte(text), &message)
	if err != nil {
		log.Err(err)
		return
	}
	log.Info().Msgf("the controller received a UI message: %+v", message)

	if message.Destination != "" {
		// client message for a private message
		c.gossiper.AddPrivateMessage(message.Contents, message.Destination, c.gossiper.GetIdentifier(), 10)
		c.messages = append(c.messages, CtrlMessage{c.identifier, 0, message.Contents})
	} else {
		// client message for regular text message
		id := c.gossiper.AddMessage(message.Contents)
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, id)
		w.Write(buf)
		c.messages = append(c.messages, CtrlMessage{c.identifier, id, message.Contents})
	}
}

// GetNode returns list of nodes as json encoded slice of string
func (c *Drone) GetNode(w http.ResponseWriter, r *http.Request) {
	hosts := c.gossiper.GetNodes()
	if err := json.NewEncoder(w).Encode(hosts); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetDirectNode returns list of nodes in the routing table as json encoded slice of string
func (c *Drone) GetDirectNode(w http.ResponseWriter, r *http.Request) {
	hosts := c.gossiper.GetDirectNodes()
	if err := json.NewEncoder(w).Encode(hosts); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PostNode with address of node in the body as a string
func (c *Drone) PostNode(w http.ResponseWriter, r *http.Request) {
	text, ok := readString(w, r)
	if !ok {
		return
	}
	log.Info().Msgf("GUI add node %s", text)
	if err := c.gossiper.AddAddresses(text); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// GetIdentifier returns the identifier as a raw string in the body
func (c *Drone) GetIdentifier(w http.ResponseWriter, r *http.Request) {
	id := c.gossiper.GetIdentifier()
	if _, err := w.Write([]byte(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Msg("GUI identifier request")
}

// SetIdentifier reads the identifier as a raw string in the body and sets the
// gossiper.
func (c *Drone) SetIdentifier(w http.ResponseWriter, r *http.Request) {
	id, ok := readString(w, r)
	if !ok {
		return
	}

	log.Info().Msg("GUI set identifier")

	c.gossiper.SetIdentifier(id)
}

// GetRoutingTable returns the routing table
func (c *Drone) GetRoutingTable(w http.ResponseWriter, r *http.Request) {
	routing := c.gossiper.GetRoutingTable()
	if err := json.NewEncoder(w).Encode(routing); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AddRoute adds a route to the gossiper
func (c *Drone) AddRoute(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Err(err).Msg("failed to parse form")
	}

	peerName := r.PostFormValue("peerName")
	if peerName == "" {
		log.Error().Msg("peerName is empty")
		return
	}

	nextHop := r.PostFormValue("nextHop")
	if nextHop == "" {
		log.Error().Msg("nextHop is empty")
		return
	}

	c.gossiper.AddRoute(peerName, nextHop)
}

// GetLocalAddr returns the gossiper's local addr
func (c *Drone) GetLocalAddr(w http.ResponseWriter, r *http.Request) {
	localAddr := c.gossiper.GetLocalAddr()

	_, err := w.Write([]byte(localAddr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HandleGossipMessage handle specific messages concerning the drone
func (c *Drone) HandleGossipMessage(origin string, msg gossip.GossipPacket) {
	//fmt.Println("DRONE message Handler")
	//fmt.Println(msg)
	//c.Lock()
	//defer c.Unlock()
	if msg.Rumor != nil {
		//fmt.Println(msg.Rumor)
		if msg.Rumor.Extra != nil {
			if msg.Rumor.Extra.SwarmInit != nil {
				go func() {
					log.Printf("%s Swarm init received", c.identifier)
					//Begin mapping phase
					target := c.targetsMapper.MapTargets(msg.Rumor.Extra.SwarmInit.InitialPos, msg.Rumor.Extra.SwarmInit.TargetPos)
					targets, _ := c.mapping.Propose(c.gossiper, msg.Rumor.Extra.SwarmInit.PatternID, target)
					c.target = targets[c.droneID]
				}()
			} else {
				//fmt.Println("New Paxos Message")
				/*if msg.Rumor.Extra.PaxosTLC != nil {
					fmt.Println("New PAXOS TLC", c.gossiper.GetIdentifier())
				}*/
				//c.naming.HandleExtraMessage(c.gossiper, msg.Rumor.Extra) // TO TEST PAXOS with naming

				//Handle Paxos
				c.mapping.HandleExtraMessage(c.gossiper, msg.Rumor.Extra)
			}
		}
		c.Lock()
		c.messages = append(c.messages, CtrlMessage{msg.Rumor.Origin, msg.Rumor.ID, msg.Rumor.Text})
		c.Unlock()
	} else if msg.Private != nil {
		c.Lock()
		c.messages = append(c.messages, CtrlMessage{msg.Private.Origin, 0, msg.Private.Text})
		c.Unlock()
	}

	//log.Info().Msgf("messages %v", c.messages)
}

func (c *Drone) GetTarget() r3.Vec {
	return c.target
}
func (c *Drone) GetDroneID() uint32 {
	return c.droneID
}

// TO TEST PAXOS with naming
func (c *Drone) ProposeName(metahash string, filename string) (string, error) {
	return c.naming.Propose(c.gossiper, metahash, filename)
}

func readString(w http.ResponseWriter, r *http.Request) (string, bool) {
	buff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read message", http.StatusBadRequest)
		return "", false
	}

	return string(buff), true
}

// logging is a utility function that logs the http server events
func logging(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Info().Str("requestID", requestID).
					Str("method", r.Method).
					Str("url", r.URL.Path).
					Str("remoteAddr", r.RemoteAddr).
					Str("agent", r.UserAgent()).Msg("")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// tracing is a utility function that adds header tracing
func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
