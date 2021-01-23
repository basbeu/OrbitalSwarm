package consensus

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

type ConsensusReader struct{}

func (c *ConsensusReader) ProposeTargets(g *gossip.Gossiper, patternID string, targets []r3.Vec) []r3.Vec {
	return nil

}
func (c *ConsensusReader) ProposePaths(g *gossip.Gossiper, patternID string, paths [][]r3.Vec) [][]r3.Vec {
	return nil
}

func (c *ConsensusReader) GetBlocks() (string, map[string]*blk.BlockContainer) {
	return "", nil
}

func (c *ConsensusReader) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {

}
