package drone

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

type pathProposition struct {
	patternID string
	paths     [][]r3.Vec
	done      chan [][]r3.Vec
}

type ConsensusClient struct {
	blockChain *paxos.BlockChain

	// PatternID -> targets
	patterns map[string][]r3.Vec
	// PatternID -> path
	paths map[string][][]r3.Vec

	mutex       sync.Mutex
	proposed    bool
	pending     []*proposition
	pendingPath []*pathProposition
}

func NewConsensusClient(numDrones, nodeIndex, paxosRetry int) *ConsensusClient {
	return &ConsensusClient{
		blockChain: paxos.NewBlockchain(numDrones, nodeIndex, paxosRetry, blk.NewGenericBlockFactory()),

		patterns: make(map[string][]r3.Vec),

		proposed:    false,
		pending:     make([]*proposition, 0),
		pendingPath: make([]*pathProposition, 0),
	}
}

func (m *ConsensusClient) ProposeTargets(g *gossip.Gossiper, patternID string, targets []r3.Vec) ([]r3.Vec, error) {
	//PatternID already mapped
	if ConsensusClient, found := m.patterns[patternID]; found {
		return ConsensusClient, nil
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

func (m *ConsensusClient) ProposePaths(g *gossip.Gossiper, patternID string, paths [][]r3.Vec) ([][]r3.Vec, error) {
	//PatternID already mapped
	if agreement, found := m.paths[patternID]; found {
		return agreement, nil
	}

	// Add the propostion to the pending list
	prop := &pathProposition{
		patternID: patternID,
		paths:     paths,
		done:      make(chan [][]r3.Vec),
	}

	m.mutex.Lock()
	m.pendingPath = append(m.pendingPath, prop)
	if !m.proposed && len(m.pending) == 0 {
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

func (m *ConsensusClient) GetBlocks() (string, map[string]*blk.BlockContainer) {
	return m.blockChain.GetBlocks()
}

func (m *ConsensusClient) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	blockContainer := m.blockChain.HandleExtraMessage(g, msg)
	if blockContainer == nil {
		return
	}

	switch blockContainer.Type {
	case blk.BlockMappingStr:
		m.handleMappingBlock(blockContainer)
	case blk.BlockPathStr:
		m.handlePathBlock(blockContainer)
	}

}

func (m *ConsensusClient) handleMappingBlock(blockContainer *blk.BlockContainer) {
	block := blockContainer.Block.(*blk.MappingBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.MappingBlockContent)

		m.patterns[blockContent.PatternID] = blockContent.Targets

		// Handle pending propositions
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.proposed = false
		pendings := make([]*proposition, 0)

		for _, p := range m.pending {
			p.done <- blockContent.Targets
			close(p.done)
		}

		m.pending = pendings
	}
}

func (m *ConsensusClient) handlePathBlock(blockContainer *blk.BlockContainer) {
	block := blockContainer.Block.(*blk.PathBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.PathBlockContent)

		m.paths[blockContent.PatternID] = blockContent.Paths

		// Handle pending propositions
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.proposed = false
		pendings := make([]*pathProposition, 0)

		for _, p := range m.pendingPath {
			p.done <- blockContent.Paths
			close(p.done)
		}

		m.pendingPath = pendings
	}
}
