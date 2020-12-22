package paxos

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
)

type TLC struct {
	paxos          *Paxos
	numParticipant int

	tlcConfirmed int
	block        extramessage.Block
}

func NewTLC(numParticipant int, nodeIndex int, paxosRetry int, blockNumber int, block extramessage.Block) *TLC {
	return &TLC{
		paxos:          NewPaxos(blockNumber, numParticipant, nodeIndex, paxosRetry),
		numParticipant: numParticipant,

		tlcConfirmed: 0,
		block:        block,
	}
}

func (t *TLC) propose(g *gossip.Gossiper, block *extramessage.Block) {
	t.paxos.propose(g, &extramessage.Block{
		BlockNumber:  block.BlockNumber,
		Filename:     block.Filename,
		Metahash:     block.Metahash,
		PreviousHash: make([]byte, 0),
	})
}

func (t *TLC) stop() {
	t.paxos.stop()
}

func (t *TLC) handleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) *extramessage.Block {
	if msg.PaxosTLC != nil {
		if msg.PaxosTLC.Block.BlockNumber == t.block.BlockNumber {
			t.tlcConfirmed++
			if t.tlcConfirmed >= t.numParticipant/2+1 {
				// log.Printf("%s Consensus of consensus !", g.identifier)
				return &msg.PaxosTLC.Block
			}
		}
	} else {
		block := t.paxos.handle(g, msg)
		if block != nil {
			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosTLC: &extramessage.PaxosTLC{
					Block: extramessage.Block{
						BlockNumber:  t.block.BlockNumber,
						PreviousHash: t.block.PreviousHash,
						Filename:     block.Filename,
						Metahash:     block.Metahash,
					},
				},
			})
		}
	}
	return nil
}
