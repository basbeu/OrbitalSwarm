package paxos

import (
	"encoding/hex"

	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
)

// BlockChain allow to handle HandlingPackets
type BlockChain struct {
	numParticipant int
	nodeIndex      int
	paxosRetry     int

	tail   blk.Block
	blocks map[string]blk.Block
	tlc    *TLC
}

func NewBlockchain(numParticipant int, nodeIndex int, paxosRetry int) *BlockChain {
	blocks := make(map[string]blk.Block)

	return &BlockChain{
		numParticipant: numParticipant,
		nodeIndex:      nodeIndex,
		paxosRetry:     paxosRetry,

		tlc: NewTLC(numParticipant, nodeIndex, paxosRetry, 0, &blk.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
		}),
		tail:   nil,
		blocks: blocks,
	}
}

func (b *BlockChain) propose(g *gossip.Gossiper, metahash []byte, filename string) {
	if b.tail == nil {
		// First block
		b.tlc.propose(g, &blk.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),

			Filename: filename,
			Metahash: metahash,
		})
	} else {
		b.tlc.propose(g, &blk.NamingBlock{
			BlockNum: b.tail.BlockNumber() + 1,
			PrevHash: b.tail.Hash(),

			Filename: filename,
			Metahash: metahash,
		})
	}
}

// GetBlocks returns all the blocks added so far. Key should be hexadecimal
// representation of the block's hash. The first return is the hexadecimal
// hash of the last block.
func (b *BlockChain) GetBlocks() (string, map[string]blk.Block) {
	if b.tail == nil {
		return hex.EncodeToString(make([]byte, 32)), b.blocks
	}
	return hex.EncodeToString(b.tail.Hash()), b.blocks
}

func (b *BlockChain) handleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) blk.Block {
	block := b.tlc.handleExtraMessage(g, msg)
	if block != nil {
		b.blocks[hex.EncodeToString(block.Hash())] = block
		b.tail = block
		b.tlc.stop()
		b.tlc = NewTLC(b.numParticipant, b.nodeIndex, b.paxosRetry, b.tail.BlockNumber()+1, &blk.NamingBlock{
			BlockNum: b.tail.BlockNumber() + 1,
			PrevHash: b.tail.Hash(),
		})
	}
	return block
}
