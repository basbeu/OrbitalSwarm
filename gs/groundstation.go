// ========== CS-438 orbitalswarm Skeleton ===========
// *** Do not change this file ***

package gs

import (
	"context"
	"fmt"
	"net"
	"net/http"
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

// GroundStation is responsible to be the glue between the gossiping protocol and
// the ui, dispatching responses and messages etc
type GroundStation struct {
	sync.Mutex
	uiAddress     string
	identifier    string
	gossipAddress string
	gossiper      gossip.BaseGossiper
	cliConn       net.Conn
	hub           *Hub

	handler chan []byte
}

// NewGroundStation returns the controller that sets up the gossiping state machine
// as well as the web routing. It uses the same gossiping address for the
// identifier.
func NewGroundStation(identifier, uiAddress, gossipAddress string, g gossip.BaseGossiper, addresses ...string) *GroundStation {
	handler := make(chan []byte)
	gs := &GroundStation{
		identifier:    identifier,
		uiAddress:     uiAddress,
		gossipAddress: gossipAddress,
		gossiper:      g,
		handler:       handler,
		hub:           newHub(handler),
	}

	g.RegisterCallback(gs.HandleGossipMessage)
	return gs
}

// Run Launch the groundstation
func (g *GroundStation) Run() {
	logger := log.With().Timestamp().Str("role", "http proxy").Logger()

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	go g.hub.run()

	r := mux.NewRouter()

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
		g.hub.broadcast <- make([]byte)
		//TODO: Handle messages
		// log.Info().Msgf("messages %v",)
	}
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
