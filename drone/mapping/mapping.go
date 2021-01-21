package mapping

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
	targets   []r3.Vec
	done      chan []r3.Vec
}

type Mapping struct {
	blockChain *paxos.BlockChain

	// PatternID -> targets
	patterns map[string][]r3.Vec

	mutex    sync.Mutex
	proposed bool
	pending  []*proposition
}

func NewMapping(numDrones, nodeIndex, paxosRetry int) *Mapping {
	return &Mapping{
		blockChain: paxos.NewBlockchain(numDrones, nodeIndex, paxosRetry, blk.NewMappingBlockFactory()),

		proposed: false,
		pending:  make([]*proposition, 0),
	}
}

func (m *Mapping) Propose(g *gossip.Gossiper, patternID string, targets []r3.Vec) ([]r3.Vec, error) {
	//PatternID already mapped
	if mapping, found := m.patterns[patternID]; found {
		return mapping, nil
	}

	// Add the propostion to the pending list
	prop := &proposition{
		patternID: patternID,
		targets:   targets,
		done:      make(chan []r3.Vec),
	}

	m.mutex.Lock()
	m.pending = append(m.pending, prop)
	if !m.proposed {
		log.Printf("Propose mapping")
		m.proposed = true
		m.blockChain.Propose(g, &blk.MappingBlockContent{
			PatternID: prop.patternID,
			Targets:   prop.targets,
		})
	}

	m.mutex.Unlock()

	return <-prop.done, nil
}

func (m *Mapping) GetBlocks() (string, map[string]*blk.BlockContainer) {
	return m.blockChain.GetBlocks()
}

func (m *Mapping) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	blockContainer := m.blockChain.HandleExtraMessage(g, msg)
	if blockContainer == nil {
		return
	}

	block := blockContainer.Block.(*blk.MappingBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.MappingBlockContent)

		m.patterns[blockContent.PatternID] = blockContent.Targets

		// Propose next pattern if any
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.proposed = false
		pendings := make([]*proposition, 0)

		for _, p := range m.pending {
			if p.patternID == blockContent.PatternID {
				p.done <- blockContent.Targets
				close(p.done)
			} else if !m.proposed {
				m.blockChain.Propose(g, &blk.MappingBlockContent{
					PatternID: p.patternID,
					Targets:   p.targets,
				})
				m.proposed = true
				pendings = append(pendings, p)
			} else {
				pendings = append(pendings, p)
			}
		}

		m.pending = pendings
	}
}
