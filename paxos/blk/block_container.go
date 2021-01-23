package blk

import (
	"encoding/json"
	"reflect"

	"golang.org/x/xerrors"
)

type BlockContainer struct {
	Block
	Type string
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
