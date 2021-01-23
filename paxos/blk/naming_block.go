package blk

import "crypto/sha256"

// NamingBlock ...
type NamingBlock struct {
	BlockNum int // not included in the hash
	PrevHash []byte

	Content BlockContent
}

type NamingBlockContent struct {
	Metahash []byte
	Filename string
}

func (c *NamingBlockContent) Hash() []byte {
	h := sha256.New()
	h.Write(c.Metahash)
	h.Write([]byte(c.Filename))
	return h.Sum(nil)
}

func (c *NamingBlockContent) Copy() BlockContent {
	return &NamingBlockContent{
		Metahash: append([]byte{}, c.Metahash...),
		Filename: c.Filename,
	}
}

func (c *NamingBlockContent) BlockType() string {
	return BlockNamingStr
}

// Hash returns the hash of a block. It doesn't take the index.
func (b *NamingBlock) Hash() []byte {
	h := sha256.New()

	h.Write(b.PrevHash)
	h.Write(b.Content.Hash())

	return h.Sum(nil)
}

// Copy performs a deep copy of a block
func (b *NamingBlock) Copy() Block {
	if b.Content == nil {
		return &NamingBlock{
			BlockNum: b.BlockNum,
			PrevHash: append([]byte{}, b.PrevHash...),
		}
	}
	return &NamingBlock{
		BlockNum: b.BlockNum,
		PrevHash: append([]byte{}, b.PrevHash...),

		Content: b.Content.Copy(),
	}
}

func (b *NamingBlock) BlockNumber() int {
	return b.BlockNum
}

func (b *NamingBlock) PreviousHash() []byte {
	return b.PrevHash
}

func (b *NamingBlock) SetPreviousHash(prevHash []byte) {
	b.PrevHash = prevHash
}

func (b *NamingBlock) GetContent() BlockContent {
	return b.Content
}

func (b *NamingBlock) SetContent(blockContent BlockContent) {
	namingContent, ok := blockContent.(*NamingBlockContent)

	if ok {
		b.Content = namingContent.Copy()
	}
}

func (b *NamingBlock) IsContentNil() bool {
	namingContent := b.Content.(*NamingBlockContent)
	return namingContent.Metahash == nil
}
