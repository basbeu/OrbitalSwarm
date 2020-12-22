package extramessage

import "crypto/sha256"

// Blockchain data structures. Feel free to move that in a separate file and/or
// package.

// Block describes the content of a block in the blockchain.
type Block struct {
	BlockNumber  int // not included in the hash
	PreviousHash []byte

	Metahash []byte
	Filename string
}

// Hash returns the hash of a block. It doesn't take the index.
func (b Block) Hash() []byte {
	h := sha256.New()

	h.Write(b.PreviousHash)

	h.Write(b.Metahash)
	h.Write([]byte(b.Filename))

	return h.Sum(nil)
}

// Copy performs a deep copy of a block
func (b Block) Copy() *Block {
	return &Block{
		BlockNumber:  b.BlockNumber,
		PreviousHash: append([]byte{}, b.PreviousHash...),

		Metahash: append([]byte{}, b.Metahash...),
		Filename: b.Filename,
	}
}
