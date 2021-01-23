package consensus

import (
	"log"
	"sync"

	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"
)

type targetProposition struct {
	patternID string
	targets   []r3.Vec
	done      chan []r3.Vec
}

type pathProposition struct {
	patternID string
	paths     [][]r3.Vec
	done      chan [][]r3.Vec
}

type ConsensusParticipant struct {
	blockChain *paxos.BlockChain

	// PatternID -> targets
	patterns map[string][]r3.Vec
	// PatternID -> path
	paths map[string][][]r3.Vec

	mutex       sync.Mutex
	proposed    bool
	pending     []*targetProposition
	pendingPath []*pathProposition
}

func NewConsensusParticipant(numDrones, nodeIndex, paxosRetry int) *ConsensusParticipant {
	return &ConsensusParticipant{
		blockChain: paxos.NewBlockchain(numDrones, nodeIndex, paxosRetry, blk.NewGenericBlockFactory()),

		patterns: make(map[string][]r3.Vec),
		paths:    make(map[string][][]r3.Vec),

		proposed:    false,
		pending:     make([]*targetProposition, 0),
		pendingPath: make([]*pathProposition, 0),
	}
}

func (c *ConsensusParticipant) ProposeTargets(g *gossip.Gossiper, patternID string, targets []r3.Vec) []r3.Vec {
	//PatternID already mapped
	if ConsensusParticipant, found := c.patterns[patternID]; found {
		return ConsensusParticipant
	}

	// Add the propostion to the pending list
	prop := &targetProposition{
		patternID: patternID,
		targets:   targets,
		done:      make(chan []r3.Vec),
	}

	c.mutex.Lock()
	c.pending = append(c.pending, prop)
	if !c.proposed {
		log.Printf("Propose mapping")
		c.proposed = true
		c.blockChain.Propose(g, &blk.MappingBlockContent{
			PatternID: prop.patternID,
			Targets:   prop.targets,
		})
	}
	c.mutex.Unlock()

	return <-prop.done
}

func (c *ConsensusParticipant) ProposePaths(g *gossip.Gossiper, patternID string, paths [][]r3.Vec) [][]r3.Vec {
	//PatternID already mapped
	if agreement, found := c.paths[patternID]; found {
		return agreement
	}

	// Add the propostion to the pending list
	prop := &pathProposition{
		patternID: patternID,
		paths:     paths,
		done:      make(chan [][]r3.Vec),
	}

	c.mutex.Lock()
	c.pendingPath = append(c.pendingPath, prop)
	if !c.proposed {
		log.Printf("Propose paths")
		c.proposed = true
		c.blockChain.Propose(g, &blk.PathBlockContent{
			PatternID: prop.patternID,
			Paths:     prop.paths,
		})
	}
	c.mutex.Unlock()

	return <-prop.done
}

func (c *ConsensusParticipant) GetBlocks() (string, map[string]*blk.BlockContainer) {
	return c.blockChain.GetBlocks()
}

func (c *ConsensusParticipant) HandleExtraMessage(g *gossip.Gossiper, msg *extramessage.ExtraMessage) {
	blockContainer := c.blockChain.HandleExtraMessage(g, msg)
	if blockContainer == nil {
		return
	}

	switch blockContainer.Type {
	case blk.BlockMappingStr:
		log.Printf("Received a mapping block")
		c.handleMappingBlock(blockContainer)
	case blk.BlockPathStr:
		log.Printf("Received a path block")
		c.handlePathBlock(blockContainer)
	}
}

func (c *ConsensusParticipant) handleMappingBlock(blockContainer *blk.BlockContainer) {
	block := blockContainer.Block.(*blk.MappingBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.MappingBlockContent)

		c.patterns[blockContent.PatternID] = blockContent.Targets

		// Handle pending propositions
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.proposed = false

		for _, p := range c.pending {
			p.done <- blockContent.Targets
			close(p.done)
		}

		c.pending = make([]*targetProposition, 0)
	}
}

func (c *ConsensusParticipant) handlePathBlock(blockContainer *blk.BlockContainer) {
	block := blockContainer.Block.(*blk.PathBlock)
	if block != nil {
		blockContent := block.GetContent().(*blk.PathBlockContent)

		c.paths[blockContent.PatternID] = blockContent.Paths

		// Handle pending propositions
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.proposed = false

		for _, p := range c.pendingPath {
			p.done <- blockContent.Paths
			close(p.done)
		}

		c.pendingPath = make([]*pathProposition, 0)
		log.Printf("Quit path handle")
	}
}
