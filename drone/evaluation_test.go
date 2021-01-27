package drone

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"go.dedis.ch/cs438/orbitalswarm/drone/mapping"

	"go.dedis.ch/cs438/orbitalswarm/drone/consensus"
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

func createSwarmTest(i, numDrones, numParticipants, antiEntropy, routeTimer, paxosRetry int) (*Swarm, []r3.Vec, *gossip.Gossiper, consensus.ConsensusClient) {
	swarm, pos := NewSwarm(numDrones, numParticipants, 2222+i, 5000+i, antiEntropy, routeTimer, paxosRetry, "127.0.0.1", "127.0.0.1")

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

	consensus := consensus.NewConsensusReader(numDrones, numDrones+1, paxosRetry)

	return swarm, pos, g, consensus
}

func TestAllDronesPaxosProposer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	trials := 10
	drones := []int{5, 7, 9, 13}
	timings := make([][]int64, len(drones))
	targetsPos := targetsPos(17)

	startingPort := 0
	for i, numDrones := range drones {
		timings[i] = make([]int64, trials)
		for j := 0; j < trials; j++ {
			paxosRetry := 3
			routeTimer := 0
			antiEntropy := 10
			//numDrones := 5

			swarm, pos, g, consensus := createSwarmTest(startingPort, numDrones, numDrones, antiEntropy, routeTimer, paxosRetry)

			targets := targetsPos[:numDrones]

			consensusReached := make(chan struct{})
			g.RegisterCallback(func(origin string, msg gossip.GossipPacket) {
				if msg.Rumor != nil && msg.Rumor.Extra != nil {
					blockContainer := consensus.HandleExtraMessage(g, msg.Rumor.Extra)
					if blockContainer != nil && blockContainer.Type == blk.BlockPathStr {
						close(consensusReached)
					}
				}
			})
			start := time.Now()
			g.AddExtraMessage(&extramessage.ExtraMessage{
				SwarmInit: &extramessage.SwarmInit{
					PatternID:  strconv.Itoa(j),
					InitialPos: pos,
					TargetPos:  targets,
				},
			})

			<-consensusReached
			timings[i][j] = time.Since(start).Nanoseconds()

			swarm.Stop()
			g.Stop()
			startingPort += numDrones
		}
	}
	fmt.Println(timings)
	f, err := os.Create("test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	var header string
	for _, drone := range drones {
		header += fmt.Sprintf("%d;", drone)
	}
	f.WriteString(header + "finish\n")
	for trial := 0; trial < trials; trial++ {
		var resultLine string
		for config := 0; config < len(drones); config++ {
			resultLine += fmt.Sprintf("%d;", timings[config][trial])
		}

		f.WriteString(resultLine + "\n")
	}
	f.Close()

}

func initialPos(num int) []r3.Vec {
	pos := make([]r3.Vec, num)

	for i := 0; i < num; i++ {
		pos[i] = r3.Vec{float64(i), float64(i), float64(i)}
	}

	return pos
}

func targetsPos(num int) []r3.Vec {
	pos := make([]r3.Vec, num)

	for i := 0; i < num; i++ {
		pos[i] = r3.Vec{float64(i), float64(i + 10), float64(i)}
	}

	return pos
}

func BenchmarkHungarianMapper(b *testing.B) {
	mapper := mapping.NewHungarianMapper()
	num := 5
	initialPos := initialPos(num)
	targetPos := targetsPos(num)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapper.MapTargets(initialPos, targetPos)
	}
}
