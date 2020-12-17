package gossip

import (
	"encoding/hex"
	"fmt"
	"sync"

	"go.dedis.ch/cs438/hw3/gossip/types"
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
		blockChain: NewBlockchain(numParticipant, nodeIndex, paxosRetry),

		files:    make(map[string]string),
		proposed: false,
		pending:  make([]*proposition, 0),
	}
}

func (n *Naming) propose(g *Gossiper, metahash string, filename string) (string, error) {
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
		n.blockChain.propose(g, hash, filename)
	}
	n.mutex.Unlock()

	return <-prop.done, nil
}

func (n *Naming) GetBlocks() (string, map[string]types.Block) {
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

func (n *Naming) handleExtraMessage(g *Gossiper, msg *types.ExtraMessage) {
	block := n.blockChain.handleExtraMessage(g, msg)
	if block != nil {
		metahash := hex.EncodeToString(block.Metahash)
		n.files[block.Filename] = hex.EncodeToString(block.Metahash)

		// Propose next file if any
		n.mutex.Lock()
		defer n.mutex.Unlock()
		n.proposed = false
		pendings := make([]*proposition, 0)

		for _, p := range n.pending {
			// resolved
			if p.metahash == metahash {
				p.done <- block.Filename
				close(p.done)
			} else if !n.proposed {
				data, err := hex.DecodeString(p.metahash)
				if err != nil {
					log.Printf("Unable to decode metahash string")
				}
				n.blockChain.propose(g, data, p.filename)
				n.proposed = true
				pendings = append(pendings, p)
			} else {
				pendings = append(pendings, p)
			}
		}
		n.pending = pendings
	}
}
