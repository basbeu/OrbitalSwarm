package gossip

import (
	"go.dedis.ch/cs438/hw3/gossip/types"
)

type TLC struct {
	paxos          *Paxos
	numParticipant int

	tlcConfirmed int
	block        types.Block
}

func NewTLC(numParticipant int, nodeIndex int, paxosRetry int, blockNumber int, block types.Block) *TLC {
	return &TLC{
		paxos:          NewPaxos(blockNumber, numParticipant, nodeIndex, paxosRetry),
		numParticipant: numParticipant,

		tlcConfirmed: 0,
		block:        block,
	}
}

func (t *TLC) propose(g *Gossiper, block *types.Block) {
	t.paxos.propose(g, &types.Block{
		BlockNumber:  block.BlockNumber,
		Filename:     block.Filename,
		Metahash:     block.Metahash,
		PreviousHash: make([]byte, 0),
	})
}

func (t *TLC) stop() {
	t.paxos.stop()
}

func (t *TLC) handleExtraMessage(g *Gossiper, msg *types.ExtraMessage) *types.Block {
	if msg.TLC != nil {
		if msg.TLC.Block.BlockNumber == t.block.BlockNumber {
			t.tlcConfirmed++
			if t.tlcConfirmed >= t.numParticipant/2+1 {
				// log.Printf("%s Consensus of consensus !", g.identifier)
				return &msg.TLC.Block
			}
		}
	} else {
		block := t.paxos.handle(g, msg)
		if block != nil {
			g.AddExtraMessage(&types.ExtraMessage{
				TLC: &types.TLC{
					Block: types.Block{
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
