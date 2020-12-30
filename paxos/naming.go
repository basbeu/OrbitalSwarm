package paxos

import (
	"encoding/hex"
	"fmt"
	"sync"

	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"go.dedis.ch/onet/v3/log"
)

type proposition struct {
	filename string
	metahash string
	done     chan string
}

// Naming base structure for a naming object
type Naming struct {
	blockChain *BlockChain

	// filename -> metahash
	files map[string]string

	mutex    sync.Mutex
	proposed bool
	pending  []*proposition
}

func NewNaming(numParticipant int, nodeIndex int, paxosRetry int) *Naming {
	return &Naming{
		blockChain: NewBlockchain(numParticipant, nodeIndex, paxosRetry, blk.NewNamingBlockFactory()),

		files:    make(map[string]string),
		proposed: false,
		pending:  make([]*proposition, 0),
	}
}

func (n *Naming) Propose(g *gossip.Gossiper, metahash string, filename string) (string, error) {
	hash, err := hex.DecodeString(metahash)
	if err != nil {
		return "", err
	}

	// Filename already existing
	if _, found := n.files[filename]; found {
		return "", fmt.Errorf("Filename already attributed")
	}

	// Consensus already existing for this metafile
	file, err := n.filenameFromMetaHash(metahash)
	if err == nil {
		return file, nil
	}

	// Add to the pending list
	prop := &proposition{
		filename: filename,
		metahash: metahash,
		done:     make(chan string, 1),
	}

	n.mutex.Lock()
	n.pending = append(n.pending, prop)
	if !n.proposed {
		log.Printf("Propose value")
		n.proposed = true
		n.blockChain.propose(g, &blk.NamingBlockContent{
			Metahash: hash,
			Filename: filename,
		})
	}
	n.mutex.Unlock()

	return <-prop.done, nil
}

func (n *Naming) GetBlocks() (string, map[string]*blk.BlockContainer) {
	return n.blockChain.GetBlocks()
}

func (n *Naming) filenameFromMetaHash(metahash string) (string, error) {
	for filename, hash := range n.files {
		if hash == metahash {
			return filename, nil
		}
	}
	return "", fmt.Errorf("Unable to find this metafile")
}

func (n *Naming) getFiles() bool {
	return false
}

func (n *Naming) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	blockContainer := n.blockChain.handleExtraMessage(g, msg)
	if blockContainer == nil {
		return
	}

	block := blockContainer.Block.(*blk.NamingBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.NamingBlockContent)
		metahash := hex.EncodeToString(blockContent.Metahash)
		n.files[blockContent.Filename] = hex.EncodeToString(blockContent.Metahash)

		// Propose next file if any
		n.mutex.Lock()
		defer n.mutex.Unlock()
		n.proposed = false
		pendings := make([]*proposition, 0)

		for _, p := range n.pending {
			// resolved
			if p.metahash == metahash {
				p.done <- blockContent.Filename
				close(p.done)
			} else if !n.proposed {
				data, err := hex.DecodeString(p.metahash)
				if err != nil {
					log.Printf("Unable to decode metahash string")
				}
				n.blockChain.propose(g, &blk.NamingBlockContent{
					Metahash: data,
					Filename: p.filename,
				})
				n.proposed = true
				pendings = append(pendings, p)
			} else {
				pendings = append(pendings, p)
			}
		}
		n.pending = pendings
	}
}
