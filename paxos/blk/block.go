package blk

const (
	BlockNamingStr  = "NamingBlock"
	BlockMappingStr = "MappingBlock"
	BlockPathStr    = "PathBlock"
)

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
	BlockType() string
}

type BlockFactory interface {
	NewEmptyBlock() *BlockContainer
	NewGenesisBlock(blockContent BlockContent) *BlockContainer
	NewBlock(blockType string, blockNumber int, previousHash []byte, content BlockContent) *BlockContainer
}
