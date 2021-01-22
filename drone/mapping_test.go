package drone

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestMapping(t *testing.T) {
	paxosRetry := 3
	routeTimer := 0
	antiEntropy := 10
	numDrones := 5

	swarm, pos := NewSwarm(numDrones, 2222, 5000, antiEntropy, routeTimer, paxosRetry, "127.0.0.1", "127.0.0.1")

	go swarm.Run()

	fac := gossip.GetFactory()
	g, err := fac.New("127.0.0.1:33000", "GS", antiEntropy, routeTimer, numDrones)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 2)

	addresses := swarm.DronesAddresses()
	g.AddAddresses(addresses...)
	ready := make(chan struct{})
	go g.Run(ready)
	<-ready

	fmt.Println("Positions :", pos)
	targets := []r3.Vec{
		r3.Vec{X: 0, Y: 10, Z: 0},
		r3.Vec{X: 0, Y: 10, Z: 2},
		r3.Vec{X: 2, Y: 10, Z: 0},
		r3.Vec{X: 2, Y: 10, Z: 2},
		r3.Vec{X: 4, Y: 10, Z: 2},
	}

	g.AddExtraMessage(&extramessage.ExtraMessage{
		SwarmInit: &extramessage.SwarmInit{
			PatternID:  "pattern1",
			InitialPos: pos,
			TargetPos:  targets,
		},
	})

	time.Sleep(time.Second * 10)

	assignments := swarm.DroneTargets()
	fmt.Println("Targets:", assignments)
	for i, assignment := range assignments {
		require.Equal(t, targets[i], assignment)
	}
}
