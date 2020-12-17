package paxos

type TLC struct {
	paxos          *Paxos
	numParticipant int

	tlcConfirmed int
	block        Block
}

func NewTLC(numParticipant int, nodeIndex int, paxosRetry int, blockNumber int, block Block) *TLC {
	return &TLC{
		paxos:          NewPaxos(blockNumber, numParticipant, nodeIndex, paxosRetry),
		numParticipant: numParticipant,

		tlcConfirmed: 0,
		block:        block,
	}
}

func (t *TLC) propose(g *Gossiper, block *Block) {
	t.paxos.propose(g, &Block{
		BlockNumber:  block.BlockNumber,
		Filename:     block.Filename,
		Metahash:     block.Metahash,
		PreviousHash: make([]byte, 0),
	})
}

func (t *TLC) stop() {
	t.paxos.stop()
}

func (t *TLC) handleExtraMessage(g *Gossiper, msg *ExtraMessage) *Block {
	if msg.paxosTLC != nil {
		if msg.paxosTLC.Block.BlockNumber == t.block.BlockNumber {
			t.tlcConfirmed++
			if t.tlcConfirmed >= t.numParticipant/2+1 {
				// log.Printf("%s Consensus of consensus !", g.identifier)
				return &msg.paxosTLC.Block
			}
		}
	} else {
		block := t.paxos.handle(g, msg)
		if block != nil {
			g.AddExtraMessage(&ExtraMessage{
				paxosTLC: &PaxosTLC{
					Block: Block{
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
