package blk

import (
	"encoding/json"
	"reflect"

	"golang.org/x/xerrors"
)

// Blockchain data structures. Feel free to move that in a separate file and/or
// package.

const (
	BlockNamingStr  = "NamingBlock"
	BlockMappingStr = "MappingBlock"
	BlockPathStr    = "PathBlock"
)

type BlockContainer struct {
	Block
	Type string
}

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

func (b *BlockContainer) UnmarshalJSON(data []byte) error {
	//Setup blocktypes
	blockTypes := map[string]reflect.Type{
		BlockNamingStr:  reflect.TypeOf(NamingBlock{}),
		BlockMappingStr: reflect.TypeOf(MappingBlock{}),
		BlockPathStr:    reflect.TypeOf(PathBlock{}),
	}
	blockContentTypes := map[string]reflect.Type{
		BlockNamingStr:  reflect.TypeOf(NamingBlockContent{}),
		BlockMappingStr: reflect.TypeOf(MappingBlockContent{}),
		BlockPathStr:    reflect.TypeOf(PathBlockContent{}),
	}

	//Unmarshall in generic map[string]interface{}
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	//Check the json and extract fields
	blockTypeInterface, typeExists := m["Type"]
	blockMapInterface, blockExists := m["Block"]

	if !typeExists || !blockExists {
		return xerrors.New("Not a valid BlockContainer")
	}
	blockType, ok := blockTypeInterface.(string)
	if !ok {
		return xerrors.New("Not a valid BlockContainer, BlockType not valid")
	}
	if blockMapInterface == nil {
		return nil
	}
	blockMap, ok := blockMapInterface.(map[string]interface{})
	if !ok {
		return xerrors.New("Not a valid BlockContainer, BlockMap not valid")
	}
	blockContentMap, ok := blockMap["Content"]
	if !ok {
		return xerrors.New("Not a valid BlockContainer, Block.Content not valid")
	}

	//Unmarshal blockContent
	blockContentJSON, err := json.Marshal(blockContentMap)
	if err != nil {
		return err
	}
	t := blockContentTypes[blockType]
	blockContent := reflect.New(t).Interface().(BlockContent)
	if err = json.Unmarshal(blockContentJSON, blockContent); err != nil {
		return err
	}

	//Unmarshal Block
	blockJSON, err := json.Marshal(blockMap)
	if err != nil {
		return err
	}
	t = blockTypes[blockType]
	block := reflect.New(t).Interface().(Block)
	json.Unmarshal(blockJSON, &block) // This method return an non-nil error because that BlockContent cannot be unmarshalled directly
	block.SetContent(blockContent)

	//Set BlockContainer attributes
	b.Type = blockType
	b.Block = block

	return nil
}

func (b *BlockContainer) Copy() *BlockContainer {
	if b.Block == nil {
		return &BlockContainer{
			Type: b.Type,
		}
	}
	return &BlockContainer{
		Type:  b.Type,
		Block: b.Block.Copy(),
	}
}

func (b *BlockContainer) IsContentNil() bool {
	return b.Block == nil || b.Block.IsContentNil()
}

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
