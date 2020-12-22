// ========== CS-438 orbitalswarm Skeleton ===========
// *** Do not change this file ***

package gs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.dedis.ch/cs438/orbitalswarm/gossip"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type key int

const (
	requestIDKey key = 0
)

// Controller is responsible to be the glue between the gossiping protocol and
// the ui, dispatching responses and messages etc
type GroundStation struct {
	sync.Mutex
	uiAddress     string
	identifier    string
	gossipAddress string
	gossiper      gossip.BaseGossiper
	cliConn       net.Conn
	messages      []CtrlMessage
	// simpleMode: true if the gossiper should broadcast messages from clients as SimpleMessages
	simpleMode bool

	HookURL *url.URL
}

type CtrlMessage struct {
	Origin string
	ID     uint32
	Text   string
}

// NewGroundStation returns the controller that sets up the gossiping state machine
// as well as the web routing. It uses the same gossiping address for the
// identifier.
func NewGroundStation(identifier, uiAddress, gossipAddress string, simpleMode bool,
	g gossip.BaseGossiper, addresses ...string) *GroundStation {

	c := &GroundStation{
		identifier:    identifier,
		uiAddress:     uiAddress,
		gossipAddress: gossipAddress,
		simpleMode:    simpleMode,
		gossiper:      g,
	}

	g.RegisterCallback(c.NewMessage)
	return c
}

// Run ...
func (c *GroundStation) Run() {
	logger := log.With().Timestamp().Str("role", "http proxy").Logger()

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	hub := newHub()
	go hub.run()

	r := mux.NewRouter()
	r.Methods("GET").Path("/message").HandlerFunc(c.GetMessage)

	r.Methods("GET").Path("/origin").HandlerFunc(c.GetDirectNode)

	r.Methods("GET").Path("/node").HandlerFunc(c.GetNode)
	r.Methods("POST").Path("/node").HandlerFunc(c.PostNode)

	r.Methods("GET").Path("/id").HandlerFunc(c.GetIdentifier)
	r.Methods("POST").Path("/id").HandlerFunc(c.SetIdentifier)

	r.Methods("GET").Path("/routing").HandlerFunc(c.GetRoutingTable)
	r.Methods("POST").Path("/routing").HandlerFunc(c.AddRoute)

	r.Methods("GET").Path("/address").HandlerFunc(c.GetLocalAddr)

	r.Methods("GET").Path("/ws").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./gs/static/")))

	server := &http.Server{
		Addr:    c.uiAddress,
		Handler: tracing(nextRequestID)(logging(logger)(r)),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

// GET /message returns all messages seen so far as json encoded Message
// XXX lot of optimizations to be done here
func (c *GroundStation) GetMessage(w http.ResponseWriter, r *http.Request) {
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

// GET /node returns list of nodes as json encoded slice of string
func (c *GroundStation) GetNode(w http.ResponseWriter, r *http.Request) {
	hosts := c.gossiper.GetNodes()
	if err := json.NewEncoder(w).Encode(hosts); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GET /origin returns list of nodes in the routing table as json encoded slice of string
func (c *GroundStation) GetDirectNode(w http.ResponseWriter, r *http.Request) {
	hosts := c.gossiper.GetDirectNodes()
	if err := json.NewEncoder(w).Encode(hosts); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST /node with address of node in the body as a string
func (c *GroundStation) PostNode(w http.ResponseWriter, r *http.Request) {
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

// GET /id returns the identifier as a raw string in the body
func (c *GroundStation) GetIdentifier(w http.ResponseWriter, r *http.Request) {
	id := c.gossiper.GetIdentifier()
	if _, err := w.Write([]byte(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info().Msg("GUI identifier request")
}

// POST /id reads the identifier as a raw string in the body and sets the
// gossiper.
func (c *GroundStation) SetIdentifier(w http.ResponseWriter, r *http.Request) {
	id, ok := readString(w, r)
	if !ok {
		return
	}

	log.Info().Msg("GUI set identifier")

	c.gossiper.SetIdentifier(id)
}

// GET /routing returns the routing table
func (c *GroundStation) GetRoutingTable(w http.ResponseWriter, r *http.Request) {
	routing := c.gossiper.GetRoutingTable()
	if err := json.NewEncoder(w).Encode(routing); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST /routing adds a route to the gossiper
func (c *GroundStation) AddRoute(w http.ResponseWriter, r *http.Request) {
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

// GET /address returns the gossiper's local addr
func (c *GroundStation) GetLocalAddr(w http.ResponseWriter, r *http.Request) {
	localAddr := c.gossiper.GetLocalAddr()

	_, err := w.Write([]byte(localAddr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewMessage ...
func (c *GroundStation) NewMessage(origin string, msg gossip.GossipPacket) {
	c.Lock()
	defer c.Unlock()

	if msg.Rumor != nil {
		c.messages = append(c.messages, CtrlMessage{msg.Rumor.Origin,
			msg.Rumor.ID, msg.Rumor.Text})
	}
	log.Info().Msgf("messages %v", c.messages)

	if c.HookURL != nil {
		cp := gossip.CallbackPacket{
			Addr: origin,
			Msg:  msg,
		}

		msgBuf, err := json.Marshal(cp)
		if err != nil {
			log.Err(err).Msg("failed to marshal packet")
			return
		}

		req := &http.Request{
			Method: "POST",
			URL:    c.HookURL,
			Header: map[string][]string{
				"Content-Type": {"application/json; charset=UTF-8"},
			},
			Body: ioutil.NopCloser(bytes.NewReader(msgBuf)),
		}

		log.Info().Msgf("sending a post callback to %s", c.HookURL)
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Err(err).Msgf("failed to call callback to %s", c.HookURL)
		}
	}
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
