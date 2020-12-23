package paxos

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
)

type ConsensusClient interface {
	Propose(g *gossip.Gossiper, metahash string, filename string) (string, error)
	GetBlocks() (string, map[string]extramessage.Block)
	HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage)
}
