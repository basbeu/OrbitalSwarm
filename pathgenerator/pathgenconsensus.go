package pathgenerator

import (
	"log"
	"sync"

	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

type proposition struct {
	patternID string
	paths     [][]r3.Vec
	done      chan [][]r3.Vec
}

type PathGen struct {
	blockChain *paxos.BlockChain

	// PatternID -> path
	patterns map[string][][]r3.Vec

	mutex    sync.Mutex
	proposed bool
	pending  []*proposition
}

func NewPathGen(numDrones, nodeIndex, paxosRetry int) *PathGen {
	return &PathGen{
		blockChain: paxos.NewBlockchain(numDrones, nodeIndex, paxosRetry, blk.NewPathBlockFactory()),

		patterns: make(map[string][][]r3.Vec),

		proposed: false,
		pending:  make([]*proposition, 0),
	}
}

func (m *PathGen) Propose(g *gossip.Gossiper, patternID string, paths [][]r3.Vec) ([][]r3.Vec, error) {
	//PatternID already mapped
	if agreement, found := m.patterns[patternID]; found {
		return agreement, nil
	}

	// Add the propostion to the pending list
	prop := &proposition{
		patternID: patternID,
		paths:     paths,
		done:      make(chan [][]r3.Vec),
	}

	m.mutex.Lock()
	m.pending = append(m.pending, prop)
	if !m.proposed {
		log.Printf("Propose mapping")
		m.proposed = true
		m.blockChain.Propose(g, &blk.PathBlockContent{
			PatternID: prop.patternID,
			Paths:     prop.paths,
		})
	}
	m.mutex.Unlock()

	return <-prop.done, nil
}

func (m *PathGen) GetBlocks() (string, map[string]*blk.BlockContainer) {
	return m.blockChain.GetBlocks()
}

func (m *PathGen) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	blockContainer := m.blockChain.HandleExtraMessage(g, msg)
	if blockContainer == nil {
		return
	}

	block := blockContainer.Block.(*blk.PathBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.PathBlockContent)

		m.patterns[blockContent.PatternID] = blockContent.Paths

		// Handle pending propositions
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.proposed = false
		pendings := make([]*proposition, 0)

		for _, p := range m.pending {
			p.done <- blockContent.Paths
			close(p.done)
		}

		m.pending = pendings
	}
}
