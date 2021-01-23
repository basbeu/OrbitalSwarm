package consensus

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

type ConsensusClient interface {
	ProposeTargets(g *gossip.Gossiper, patternID string, targets []r3.Vec) []r3.Vec
	ProposePaths(g *gossip.Gossiper, patternID string, paths [][]r3.Vec) [][]r3.Vec

	GetBlocks() (string, map[string]*blk.BlockContainer)
	HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) *blk.BlockContainer

	IsProposer() bool
}
