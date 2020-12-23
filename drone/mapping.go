package drone

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/utils"
)

type mapping struct {
}

func newMapping() *mapping {
	return &mapping{}
}

func (m *mapping) Propose(g *gossip.Gossiper, targets map[string]utils.Vec3d) (string, error) {
	//TODO
	return "", nil
}

func (m *mapping) GetBlocks() (string, map[string]extramessage.Block) {
	//TODO
	return "", make(map[string]extramessage.Block)
}

func (m *mapping) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	//TODO
}
