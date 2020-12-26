package paxos

import (
	"encoding/hex"

	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
)

// BlockChain allow to handle HandlingPackets
type BlockChain struct {
	numParticipant int
	nodeIndex      int
	paxosRetry     int

	tail   extramessage.Block
	blocks map[string]extramessage.Block
	tlc    *TLC
}

func NewBlockchain(numParticipant int, nodeIndex int, paxosRetry int) *BlockChain {
	blocks := make(map[string]extramessage.Block)

	return &BlockChain{
		numParticipant: numParticipant,
		nodeIndex:      nodeIndex,
		paxosRetry:     paxosRetry,

		tlc: NewTLC(numParticipant, nodeIndex, paxosRetry, 0, &extramessage.NamingBlock{
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
		b.tlc.propose(g, &extramessage.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),

			Filename: filename,
			Metahash: metahash,
		})
	} else {
		b.tlc.propose(g, &extramessage.NamingBlock{
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
func (b *BlockChain) GetBlocks() (string, map[string]extramessage.Block) {
	if b.tail == nil {
		return hex.EncodeToString(make([]byte, 32)), b.blocks
	}
	return hex.EncodeToString(b.tail.Hash()), b.blocks
}

func (b *BlockChain) handleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) extramessage.Block {
	block := b.tlc.handleExtraMessage(g, msg)
	if block != nil {
		b.blocks[hex.EncodeToString(block.Hash())] = block
		b.tail = block
		b.tlc.stop()
		b.tlc = NewTLC(b.numParticipant, b.nodeIndex, b.paxosRetry, b.tail.BlockNumber()+1, &extramessage.NamingBlock{
			BlockNum: b.tail.BlockNumber() + 1,
			PrevHash: b.tail.Hash(),
		})
	}
	return block
}
