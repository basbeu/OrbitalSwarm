package mapping

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

type Mapping struct {
	blockChain *paxos.BlockChain
}

func NewMapping(numDrones, nodeIndex, paxosRetry int) *Mapping {
	return &Mapping{
		blockChain: paxos.NewBlockchain(numDrones, nodeIndex, paxosRetry, blk.NewMappingBlockFactory()),
	}
}

func (m *Mapping) Propose(g *gossip.Gossiper, targets []r3.Vec) (string, error) {
	//TODO
	return "", nil
}

func (m *Mapping) GetBlocks() (string, map[string]blk.Block) {
	//TODO
	return "", make(map[string]blk.Block)
}

func (m *Mapping) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	//TODO
}
