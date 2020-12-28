// ========== CS-438 HW0 Skeleton ===========

// This file should be the entering point to your program.
// Here, we only parse the input and start the logic implemented
// in other files.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.dedis.ch/cs438/orbitalswarm/drone"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/gs"
)

const defaultGossipAddr = "127.0.0.1:33000" // IP address:port number for gossiping
const defaultName = "peerXYZ"               // Give a unique default name
const defaultPaxosRetry = 3
const defaultUIPort = "12000" // Default port number

var (
	// defaultLevel can be changed to set the desired level of the logger
	defaultLevel = zerolog.WarnLevel

	// logout is the logger configuration
	logout = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Logger is a globally available logger.
	Logger = zerolog.New(logout).Level(defaultLevel).
		With().Timestamp().Logger().
		With().Caller().Logger()
)

func main() {
	UIPort := flag.String("UIPort", defaultUIPort, "port for gossip communication with peers")
	antiEntropy := flag.Int("antiEntropy", 10, "timeout in seconds for anti-entropy (relevant only fo rPart2)' default value 10 seconds.")
	routeTimer := flag.Int("rtimer", 0, "route rumors sending period in seconds, 0 to disable sending of route rumors (default)")
	numParticipants := flag.Int("numParticipants", -1, "number of participants in the Paxos consensus box.")
	nodeIndex := flag.Int("nodeIndex", -1, "index of the node with respect to all the participants")
	paxosRetry := flag.Int("paxosRetry", defaultPaxosRetry, "number of seconds a Paxos proposer waits until retrying")

	flag.Parse()

	if *numParticipants < 0 {
		Logger.Error().Msg("please specify a number of participants with --numParticipants")
		return
	}

	if *nodeIndex < 0 {
		Logger.Error().Msg("please specify a node index --nodeIndex")
		return
	}

	// Generate address for the groundStation
	gossipAddress := ""
	fac := gossip.GetFactory()
	g, err := fac.New(gossipAddress, "GS", *antiEntropy, *routeTimer,
		*numParticipants, *nodeIndex, *paxosRetry)
	if err != nil {
		panic(err)
	}

	// controller := NewController(*ownName, UIAddress, gossipAddress, g, bootstrapAddr...)
	swarm, locations := drone.NewSwarm(9, 2222, 5000, *antiEntropy, *routeTimer, *paxosRetry, "127.0.0.1", "127.0.0.1")

	addresses := swarm.DronesAddresses()
	g.AddAddresses(addresses...)

	groundStation := gs.NewGroundStation("GS", "127.0.0.1:"+*UIPort, gossipAddress, g, locations)

	go swarm.Run()
	groundStation.Run()
}

func sendWatch(p gossip.CallbackPacket, u *url.URL) {
	msgBuf, err := json.Marshal(p)
	if err != nil {
		Logger.Err(err).Msg("failed to marshal packet: %v")
		return
	}

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			"Content-Type": {"application/json; charset=UTF-8"},
		},
		Body: ioutil.NopCloser(bytes.NewReader(msgBuf)),
	}

	Logger.Info().Msgf("sending a post watch to %s", u)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		Logger.Err(err).Msgf("failed to call watch to %s", u)
	}
}
