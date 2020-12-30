package mapping

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

type Mapping struct {
}

func NewMapping() *Mapping {
	return &Mapping{}
}

func (m *Mapping) Propose(g *gossip.Gossiper, targets map[string]r3.Vec) (string, error) {
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
