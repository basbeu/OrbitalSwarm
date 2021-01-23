package blk

type GenericBlockFactory struct{}

func (f GenericBlockFactory) NewEmptyBlock() *BlockContainer {
	return &BlockContainer{
		Type:  BlockMappingStr,
		Block: nil,
	}
}

func (f GenericBlockFactory) NewGenesisBlock(blockContent BlockContent) *BlockContainer {
	return &BlockContainer{
		Type: BlockMappingStr,
		Block: &MappingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
			Content:  blockContent,
		},
	}
}

func (f GenericBlockFactory) NewBlock(blockType string, blockNumber int, previousHash []byte, content BlockContent) *BlockContainer {
	switch blockType {
	case BlockMappingStr:
		return &BlockContainer{
			Type: BlockMappingStr,
			Block: &MappingBlock{
				BlockNum: blockNumber,
				PrevHash: previousHash,
				Content:  content,
			},
		}
	case BlockNamingStr:
		return &BlockContainer{
			Type: BlockNamingStr,
			Block: &NamingBlock{
				BlockNum: blockNumber,
				PrevHash: previousHash,
				Content:  content,
			},
		}
	case BlockPathStr:
		return &BlockContainer{
			Type: BlockPathStr,
			Block: &PathBlock{
				BlockNum: blockNumber,
				PrevHash: previousHash,
				Content:  content,
			},
		}
	default:
		panic("Unknown type of blocks")
	}
}

func NewGenericBlockFactory() GenericBlockFactory {
	return GenericBlockFactory{}
}
