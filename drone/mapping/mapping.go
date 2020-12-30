package mapping

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"go.dedis.ch/cs438/orbitalswarm/utils"
)

type Mapping struct {
}

func NewMapping() *Mapping {
	return &Mapping{}
}

func (m *Mapping) Propose(g *gossip.Gossiper, targets map[string]utils.Vec3d) (string, error) {
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
