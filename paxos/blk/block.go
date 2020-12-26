package blk

// Blockchain data structures. Feel free to move that in a separate file and/or
// package.

// Block describes the content of a block in the blockchain.
type Block interface {
	Hash() []byte
	Copy() Block
	BlockNumber() int
	PreviousHash() []byte
	SetPreviousHash(prevHash []byte)
	GetContent() BlockContent
	SetContent(blockContent BlockContent)
	IsContentNil() bool
}

type BlockContent interface {
	Hash() []byte
	Copy() BlockContent
}

type BlockFactory interface {
	NewFirstBlock(blockContent BlockContent) Block
	NewBlock(blockNumber int, previousHash []byte, content BlockContent) Block
}
