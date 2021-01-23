package paxos

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"go.dedis.ch/onet/v3/log"
)

type TLC struct {
	paxos          *Paxos
	numParticipant int
	blockNumber    int

	tlcConfirmed int
	block        *blk.BlockContainer
}

func NewTLC(numParticipant int, nodeIndex int, paxosRetry int, blockNumber int, blockFactory blk.BlockFactory) *TLC {
	return &TLC{
		paxos:          NewPaxos(blockNumber, numParticipant, nodeIndex, paxosRetry, blockFactory),
		numParticipant: numParticipant,
		blockNumber:    blockNumber,

		tlcConfirmed: 0,
	}
}

func (t *TLC) propose(g *gossip.Gossiper, block *blk.BlockContainer) {
	t.block = block
	t.paxos.propose(g, block)
}

func (t *TLC) stop() {
	t.paxos.stop()
}

func (t *TLC) handleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) *blk.BlockContainer {
	if msg.PaxosTLC != nil {
		if msg.PaxosTLC.Value.BlockNumber() == t.blockNumber {
			// log.Printf("%s Consensus call ! %d", g.GetIdentifier(), t.block.BlockNumber())
			t.tlcConfirmed++
			if t.tlcConfirmed >= t.numParticipant/2+1 {
				log.Printf("%s Consensus of consensus !", g.GetIdentifier())
				return msg.PaxosTLC.Value
			}
		}
	} else {
		block := t.paxos.handle(g, msg)

		if block != nil {
			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosTLC: &extramessage.PaxosTLC{
					Value: block,
				},
			})
		}
	}
	return nil
}
