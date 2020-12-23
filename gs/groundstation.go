// ========== CS-438 orbitalswarm Skeleton ===========
// *** Do not change this file ***

package gs

import (
	"context"
	"fmt"
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

	HookURL *url.URL

	handler chan []byte
}

// NewGroundStation returns the controller that sets up the gossiping state machine
// as well as the web routing. It uses the same gossiping address for the
// identifier.
func NewGroundStation(identifier, uiAddress, gossipAddress string, g gossip.BaseGossiper, addresses ...string) *GroundStation {

	gs := &GroundStation{
		identifier:    identifier,
		uiAddress:     uiAddress,
		gossipAddress: gossipAddress,
		gossiper:      g,
		handler:       make(chan []byte),
	}

	g.RegisterCallback(gs.HandleGossipMessage)
	return gs
}

// Run ...
func (g *GroundStation) Run() {
	logger := log.With().Timestamp().Str("role", "http proxy").Logger()

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	hub := newHub(g.handler)
	go hub.run()

	r := mux.NewRouter()
	// r.Methods("GET").Path("/message").HandlerFunc(g.GetMessage)

	// r.Methods("GET").Path("/origin").HandlerFunc(g.GetDirectNode)

	// r.Methods("GET").Path("/node").HandlerFunc(g.GetNode)
	// r.Methods("POST").Path("/node").HandlerFunc(g.PostNode)

	// r.Methods("GET").Path("/id").HandlerFunc(g.GetIdentifier)
	// r.Methods("POST").Path("/id").HandlerFunc(g.SetIdentifier)

	// r.Methods("GET").Path("/routing").HandlerFunc(g.GetRoutingTable)
	// r.Methods("POST").Path("/routing").HandlerFunc(g.AddRoute)

	// r.Methods("GET").Path("/address").HandlerFunc(g.GetLocalAddr)

	r.Methods("GET").Path("/ws").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./gs/static/")))

	server := &http.Server{
		Addr:    g.uiAddress,
		Handler: tracing(nextRequestID)(logging(logger)(r)),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

// HandleGossipMessage handle gossip messages
func (g *GroundStation) HandleGossipMessage(origin string, msg gossip.GossipPacket) {
	g.Lock()
	defer g.Unlock()

	// TODO: Define what messages are important and how to handle them

	// In case of other type of message
	if msg.Rumor != nil {
		//TODO: Handle messages
		// log.Info().Msgf("messages %v",)
	}

	// if g.HookURL != nil {
	// 	cp := gossip.CallbackPacket{
	// 		Addr: origin,
	// 		Msg:  msg,
	// 	}

	// 	msgBuf, err := json.Marshal(cp)
	// 	if err != nil {
	// 		log.Err(err).Msg("failed to marshal packet")
	// 		return
	// 	}

	// 	req := &http.Request{
	// 		Method: "POST",
	// 		URL:    g.HookURL,
	// 		Header: map[string][]string{
	// 			"Content-Type": {"application/json; charset=UTF-8"},
	// 		},
	// 		Body: ioutil.NopCloser(bytes.NewReader(msgBuf)),
	// 	}

	// 	log.Info().Msgf("sending a post callback to %s", g.HookURL)
	// 	_, err = http.DefaultClient.Do(req)
	// 	if err != nil {
	// 		log.Err(err).Msgf("failed to call callback to %s", g.HookURL)
	// 	}
	// }
}

// // GET /message returns all messages seen so far as json encoded Message
// // XXX lot of optimizations to be done here
// func (g *GroundStation) GetMessage(w http.ResponseWriter, r *http.Request) {
// 	g.Lock()
// 	defer g.Unlock()
// 	log.Info().Msgf("These are the msgs %v", g.messages)
// 	if err := json.NewEncoder(w).Encode(g.messages); err != nil {
// 		log.Err(err)
// 		http.Error(w, "could not encode json", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Info().Msg("GUI request for the messages received by the gossiper")
// }

// // GET /node returns list of nodes as json encoded slice of string
// func (g *GroundStation) GetNode(w http.ResponseWriter, r *http.Request) {
// 	hosts := g.gossiper.GetNodes()
// 	if err := json.NewEncoder(w).Encode(hosts); err != nil {
// 		log.Err(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// // GET /origin returns list of nodes in the routing table as json encoded slice of string
// func (g *GroundStation) GetDirectNode(w http.ResponseWriter, r *http.Request) {
// 	hosts := g.gossiper.GetDirectNodes()
// 	if err := json.NewEncoder(w).Encode(hosts); err != nil {
// 		log.Err(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// // POST /node with address of node in the body as a string
// func (g *GroundStation) PostNode(w http.ResponseWriter, r *http.Request) {
// 	text, ok := readString(w, r)
// 	if !ok {
// 		return
// 	}
// 	log.Info().Msgf("GUI add node %s", text)
// 	if err := g.gossiper.AddAddresses(text); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// }

// // GET /id returns the identifier as a raw string in the body
// func (g *GroundStation) GetIdentifier(w http.ResponseWriter, r *http.Request) {
// 	id := g.gossiper.GetIdentifier()
// 	if _, err := w.Write([]byte(id)); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	log.Info().Msg("GUI identifier request")
// }

// // POST /id reads the identifier as a raw string in the body and sets the
// // gossiper.
// func (g *GroundStation) SetIdentifier(w http.ResponseWriter, r *http.Request) {
// 	id, ok := readString(w, r)
// 	if !ok {
// 		return
// 	}

// 	log.Info().Msg("GUI set identifier")

// 	g.gossiper.SetIdentifier(id)
// }

// // GET /routing returns the routing table
// func (g *GroundStation) GetRoutingTable(w http.ResponseWriter, r *http.Request) {
// 	routing := g.gossiper.GetRoutingTable()
// 	if err := json.NewEncoder(w).Encode(routing); err != nil {
// 		log.Err(err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// // POST /routing adds a route to the gossiper
// func (g *GroundStation) AddRoute(w http.ResponseWriter, r *http.Request) {
// 	err := r.ParseForm()
// 	if err != nil {
// 		log.Err(err).Msg("failed to parse form")
// 	}

// 	peerName := r.PostFormValue("peerName")
// 	if peerName == "" {
// 		log.Error().Msg("peerName is empty")
// 		return
// 	}

// 	nextHop := r.PostFormValue("nextHop")
// 	if nextHop == "" {
// 		log.Error().Msg("nextHop is empty")
// 		return
// 	}

// 	g.gossiper.AddRoute(peerName, nextHop)
// }

// // GET /address returns the gossiper's local addr
// func (g *GroundStation) GetLocalAddr(w http.ResponseWriter, r *http.Request) {
// 	localAddr := g.gossiper.GetLocalAddr()

// 	_, err := w.Write([]byte(localAddr))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

// func readString(w http.ResponseWriter, r *http.Request) (string, bool) {
// 	buff, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "could not read message", http.StatusBadRequest)
// 		return "", false
// 	}

// 	return string(buff), true
// }

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
