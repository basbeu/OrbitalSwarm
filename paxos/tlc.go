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

	tlcConfirmed int
	block        *blk.BlockContainer
}

func NewTLC(numParticipant int, nodeIndex int, paxosRetry int, blockNumber int, block *blk.BlockContainer, blockFactory blk.BlockFactory) *TLC {
	return &TLC{
		paxos:          NewPaxos(blockNumber, numParticipant, nodeIndex, paxosRetry, blockFactory),
		numParticipant: numParticipant,

		tlcConfirmed: 0,
		block:        block,
	}
}

func (t *TLC) propose(g *gossip.Gossiper, block *blk.BlockContainer) {
	proposed := block.Copy()
	proposed.SetPreviousHash(make([]byte, 0))
	t.paxos.propose(g, proposed)
}

func (t *TLC) stop() {
	t.paxos.stop()
}

func (t *TLC) handleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) *blk.BlockContainer {
	if msg.PaxosTLC != nil {
		if msg.PaxosTLC.Value.BlockNumber() == t.block.BlockNumber() {
			log.Printf("%s Consensus call ! %d", g.GetIdentifier(), t.block.BlockNumber())
			t.tlcConfirmed++
			if t.tlcConfirmed >= t.numParticipant/2+1 {
				log.Printf("%s Consensus of consensus !", g.GetIdentifier())
				return msg.PaxosTLC.Value
			}
		}
	} else {
		block := t.paxos.handle(g, msg)

		if block != nil {
			blockTLC := t.block.Copy()
			blockTLC.SetContent(block.GetContent())

			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosTLC: &extramessage.PaxosTLC{
					Value: blockTLC,
				},
			})
		}
	}
	return nil
}
