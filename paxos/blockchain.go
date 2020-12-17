package paxos

import (
	"encoding/hex"
)

// BlockChain allow to handle HandlingPackets
type BlockChain struct {
	numParticipant int
	nodeIndex      int
	paxosRetry     int

	tail   *Block
	blocks map[string]Block
	tlc    *TLC
}

func NewBlockchain(numParticipant int, nodeIndex int, paxosRetry int) *BlockChain {
	blocks := make(map[string]Block)

	return &BlockChain{
		numParticipant: numParticipant,
		nodeIndex:      nodeIndex,
		paxosRetry:     paxosRetry,

		tlc: NewTLC(numParticipant, nodeIndex, paxosRetry, 0, Block{
			BlockNumber:  0,
			PreviousHash: make([]byte, 32),
		}),
		tail:   nil,
		blocks: blocks,
	}
}

func (b *BlockChain) propose(g *Gossiper, metahash []byte, filename string) {
	if b.tail == nil {
		// First block
		b.tlc.propose(g, &Block{
			BlockNumber:  0,
			PreviousHash: make([]byte, 32),

			Filename: filename,
			Metahash: metahash,
		})
	} else {
		b.tlc.propose(g, &Block{
			BlockNumber:  b.tail.BlockNumber + 1,
			PreviousHash: b.tail.Hash(),

			Filename: filename,
			Metahash: metahash,
		})
	}
}

// GetBlocks returns all the blocks added so far. Key should be hexadecimal
// representation of the block's hash. The first return is the hexadecimal
// hash of the last block.
func (b *BlockChain) GetBlocks() (string, map[string]Block) {
	if b.tail == nil {
		return hex.EncodeToString(make([]byte, 32)), b.blocks
	}
	return hex.EncodeToString(b.tail.Hash()), b.blocks
}

func (b *BlockChain) handleExtraMessage(g *Gossiper, msg *ExtraMessage) *Block {
	block := b.tlc.handleExtraMessage(g, msg)
	if block != nil {
		b.blocks[hex.EncodeToString(block.Hash())] = *block
		b.tail = block
		b.tlc.stop()
		b.tlc = NewTLC(b.numParticipant, b.nodeIndex, b.paxosRetry, b.tail.BlockNumber+1, Block{
			BlockNumber:  b.tail.BlockNumber + 1,
			PreviousHash: b.tail.Hash(),
		})
	}
	return block
}
