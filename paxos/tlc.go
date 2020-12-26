package paxos

import (
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
)

type TLC struct {
	paxos          *Paxos
	numParticipant int

	tlcConfirmed int
	block        blk.Block
}

func NewTLC(numParticipant int, nodeIndex int, paxosRetry int, blockNumber int, block blk.Block) *TLC {
	return &TLC{
		paxos:          NewPaxos(blockNumber, numParticipant, nodeIndex, paxosRetry),
		numParticipant: numParticipant,

		tlcConfirmed: 0,
		block:        block,
	}
}

func (t *TLC) propose(g *gossip.Gossiper, block blk.Block) {
	proposed := block.Copy()
	proposed.SetPreviousHash(make([]byte, 0))
	t.paxos.propose(g, proposed)
	/*t.paxos.propose(g, &extramessage.NamingBlock{
		BlockNum: block.BlockNumber(),
		Filename: block.Filename,
		Metahash: block.Metahash,
		PrevHash: make([]byte, 0),
	})*/
}

func (t *TLC) stop() {
	t.paxos.stop()
}

func (t *TLC) handleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) blk.Block {
	if msg.PaxosTLC != nil {
		if msg.PaxosTLC.Block.BlockNumber() == t.block.BlockNumber() {
			t.tlcConfirmed++
			if t.tlcConfirmed >= t.numParticipant/2+1 {
				// log.Printf("%s Consensus of consensus !", g.identifier)
				return msg.PaxosTLC.Block
			}
		}
	} else {
		block := t.paxos.handle(g, msg)
		if block != nil {
			blockTLC := t.block.Copy()
			blockTLC.SetContent(block)

			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosTLC: &extramessage.PaxosTLC{
					Block: blockTLC,
					/*&extramessage.NamingBlock{
						BlockNum: t.block.BlockNumber(),
						PrevHash: t.block.PreviousHash(),
						Filename: block.Filename,
						Metahash: block.Metahash,
					},*/
				},
			})
		}
	}
	return nil
}
