package extramessage

import "crypto/sha256"

// Blockchain data structures. Feel free to move that in a separate file and/or
// package.

// Block describes the content of a block in the blockchain.
type Block interface {
	Hash() []byte
	Copy() Block
	BlockNumber() int
	PreviousHash() []byte
	SetPreviousHash(prevHash []byte)
	SetContent(block Block)
	IsContentNil() bool
}

// NamingBlock ...
type NamingBlock struct {
	BlockNum int // not included in the hash
	PrevHash []byte

	Metahash []byte
	Filename string
}

// Hash returns the hash of a block. It doesn't take the index.
func (b *NamingBlock) Hash() []byte {
	h := sha256.New()

	h.Write(b.PrevHash)

	h.Write(b.Metahash)
	h.Write([]byte(b.Filename))

	return h.Sum(nil)
}

// Copy performs a deep copy of a block
func (b *NamingBlock) Copy() Block {
	return &NamingBlock{
		BlockNum: b.BlockNum,
		PrevHash: append([]byte{}, b.PrevHash...),

		Metahash: append([]byte{}, b.Metahash...),
		Filename: b.Filename,
	}
}

func (b *NamingBlock) BlockNumber() int {
	return b.BlockNum
}

func (b *NamingBlock) PreviousHash() []byte {
	return b.PrevHash
}

func (b *NamingBlock) SetPreviousHash(prevHash []byte) {
	b.PrevHash = prevHash
}

func (b *NamingBlock) SetContent(block Block) {
	namingBlock, ok := block.(*NamingBlock)

	if ok {
		b.Filename = namingBlock.Filename
		b.Metahash = namingBlock.Metahash
	}
}
func (b *NamingBlock) IsContentNil() bool {
	return b.Metahash == nil
}
