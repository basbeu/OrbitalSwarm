package blk

import (
	"encoding/json"
	"reflect"

	"golang.org/x/xerrors"
)

// Blockchain data structures. Feel free to move that in a separate file and/or
// package.

const (
	blockNamingStr = "NamingBlock"
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
}

type BlockFactory interface {
	NewEmptyBlock() *BlockContainer
	NewFirstBlock(blockContent BlockContent) *BlockContainer
	NewBlock(blockNumber int, previousHash []byte, content BlockContent) *BlockContainer
}

func (b *BlockContainer) UnmarshalJSON(data []byte) error {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

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

	blockContentMap := blockMap["Content"]
	blockContentJSON, _ := json.Marshal(blockContentMap)
	blockContent := reflect.New(reflect.TypeOf(NamingBlockContent{})).Interface().(BlockContent)
	json.Unmarshal(blockContentJSON, &blockContent)

	blockJSON, _ := json.Marshal(blockMap)
	block := NamingBlock{}
	json.Unmarshal(blockJSON, &block)
	block.Content = blockContent

	b.Type = blockType
	b.Block = &block

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
