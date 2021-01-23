package blk

import (
	"crypto/sha256"
	"fmt"

	"gonum.org/v1/gonum/spatial/r3"
)

type PathBlock struct {
	BlockNum int // not included in the hash
	PrevHash []byte

	Content BlockContent
}

type PathBlockContent struct {
	PatternID string
	Paths     [][]r3.Vec
}

func (c *PathBlockContent) Hash() []byte {
	h := sha256.New()
	h.Write([]byte(c.PatternID))

	for _, path := range c.Paths {
		for _, step := range path {
			h.Write([]byte(fmt.Sprintf("%v", step)))
		}
	}

	return h.Sum(nil)
}

func (c *PathBlockContent) Copy() BlockContent {
	pathCopy := [][]r3.Vec{}

	for i, path := range c.Paths {
		pathCopy = append(pathCopy, []r3.Vec{})
		for _, step := range path {
			pathCopy[i] = append(pathCopy[i], step)
		}
	}

	return &PathBlockContent{
		PatternID: c.PatternID,
		Paths:     pathCopy,
	}
}

func (c *PathBlockContent) BlockType() string {
	return BlockPathStr
}

func (b *PathBlock) Hash() []byte {
	h := sha256.New()

	h.Write(b.PrevHash)
	h.Write(b.Content.Hash())

	return h.Sum(nil)
}
func (b *PathBlock) Copy() Block {
	if b.Content == nil {
		return &PathBlock{
			BlockNum: b.BlockNum,
			PrevHash: append([]byte{}, b.PrevHash...),
		}
	}
	return &PathBlock{
		BlockNum: b.BlockNum,
		PrevHash: append([]byte{}, b.PrevHash...),

		Content: b.Content.Copy(),
	}
}
func (b *PathBlock) BlockNumber() int {
	return b.BlockNum
}
func (b *PathBlock) PreviousHash() []byte {
	return b.PrevHash
}
func (b *PathBlock) SetPreviousHash(prevHash []byte) {
	b.PrevHash = prevHash
}
func (b *PathBlock) GetContent() BlockContent {
	return b.Content
}
func (b *PathBlock) SetContent(blockContent BlockContent) {
	pathContent, ok := blockContent.(*PathBlockContent)

	if ok {
		b.Content = pathContent.Copy()
	}
}
func (b *PathBlock) IsContentNil() bool {
	pathContent := b.Content.(*PathBlockContent)
	return pathContent.Paths == nil
}

/*type PathBlockFactory struct{}

func (f PathBlockFactory) NewEmptyBlock() *BlockContainer {
	return &BlockContainer{
		Type:  BlockPathStr,
		Block: nil,
	}
}
func (f PathBlockFactory) NewGenesisBlock(blockContent BlockContent) *BlockContainer {
	return &BlockContainer{
		Type: BlockPathStr,
		Block: &PathBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
			Content:  blockContent,
		},
	}
}
func (f PathBlockFactory) NewBlock(blockNumber int, previousHash []byte, content BlockContent) *BlockContainer {
	return &BlockContainer{
		Type: BlockPathStr,
		Block: &PathBlock{
			BlockNum: blockNumber,
			PrevHash: previousHash,
			Content:  content,
		},
	}
}

func NewPathBlockFactory() PathBlockFactory {
	return PathBlockFactory{}
}*/
