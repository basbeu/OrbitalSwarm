package blk

type MappingBlock struct {
	//TODO
	BlockNum int // not included in the hash
	PrevHash []byte

	Content BlockContent
}

type MappingBlockContent struct {
	//TODO
}

func (c *MappingBlockContent) Hash() []byte {
	//TODO
	return []byte{}
}

func (c *MappingBlockContent) Copy() BlockContent {
	//TODO
	return &MappingBlockContent{}
}

func (b *MappingBlock) Hash() []byte {
	//TODO
	return []byte{}
}

func (b *MappingBlock) Copy() Block {
	//TODO
	return &MappingBlock{}
}

func (b *MappingBlock) BlockNumber() int {
	return b.BlockNum
}

func (b *MappingBlock) PreviousHash() []byte {
	return b.PrevHash
}

func (b *MappingBlock) SetPreviousHash(prevHash []byte) {
	b.PrevHash = prevHash
}

func (b *MappingBlock) GetContent() BlockContent {
	return b.Content
}

func (b *MappingBlock) SetContent(blockContent BlockContent) {
	mappingContent, ok := blockContent.(*MappingBlockContent)

	if ok {
		b.Content = mappingContent.Copy()
	}
}

func (b *MappingBlock) IsContentNil() bool {
	//mappingContent := b.Content.(*MappingBlockContent)
	//TODO
	return false
	//return mappingContent.Metahash == nil
}

type MappingBlockFactory struct{}

func (f MappingBlockFactory) NewEmptyBlock() *BlockContainer {
	//TODO
	return &BlockContainer{}
}

func (f MappingBlockFactory) NewFirstBlock(blockContent BlockContent) *BlockContainer {
	//TODO
	return &BlockContainer{}
}

func (f MappingBlockFactory) NewBlock(blockNumber int, previousHash []byte, content BlockContent) *BlockContainer {
	//TODO
	return &BlockContainer{}
}

func NewMappingBlockFactory() MappingBlockFactory {
	return MappingBlockFactory{}
}
