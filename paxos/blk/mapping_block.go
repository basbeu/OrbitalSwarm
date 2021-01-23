package blk

import (
	"crypto/sha256"
	"fmt"

	"gonum.org/v1/gonum/spatial/r3"
)

type MappingBlock struct {
	BlockNum int // not included in the hash
	PrevHash []byte

	Content BlockContent
}

type MappingBlockContent struct {
	PatternID string
	Targets   []r3.Vec
}

func (c *MappingBlockContent) Hash() []byte {
	h := sha256.New()
	h.Write([]byte(c.PatternID))

	for _, t := range c.Targets {
		h.Write([]byte(fmt.Sprintf("%v", t)))
	}

	return h.Sum(nil)
}

func (c *MappingBlockContent) Copy() BlockContent {
	return &MappingBlockContent{
		PatternID: c.PatternID,
		Targets:   append([]r3.Vec{}, c.Targets...),
	}
}

func (c *MappingBlockContent) BlockType() string {
	return BlockMappingStr
}

func (b *MappingBlock) Hash() []byte {
	h := sha256.New()

	h.Write(b.PrevHash)
	h.Write(b.Content.Hash())

	return h.Sum(nil)
}

func (b *MappingBlock) Copy() Block {
	if b.Content == nil {
		return &MappingBlock{
			BlockNum: b.BlockNum,
			PrevHash: append([]byte{}, b.PrevHash...),
		}
	}
	return &MappingBlock{
		BlockNum: b.BlockNum,
		PrevHash: append([]byte{}, b.PrevHash...),

		Content: b.Content.Copy(),
	}
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
	mappingContent := b.Content.(*MappingBlockContent)
	return mappingContent.Targets == nil
}
