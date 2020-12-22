package paxos

import (
	"time"

	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/onet/v3/log"
)

const (
	stateNoProposal   = 1
	stateAwaitPromise = 2
	stateAwaitAccept  = 3
	stateConsensus    = 4
)

// Paxos data structure
type Paxos struct {
	// base config
	paxosSequenceID int
	nodeIndex       int
	numParticipant  int
	paxosRetry      int

	// Round Tracking
	idGenerator UniqueIDGenerator
	ID          int

	// Proposal
	proposedID   int
	state        int
	count        int
	value        *extramessage.Block
	chanMajority chan bool

	// chain BlockChain
	latestPrepareID     int
	latestAcceptedID    int
	acceptedCount       int
	latestAcceptedValue extramessage.Block

	learnerCount int
	learnerData  map[int]int

	// stop
	chanEnd chan bool
}

// NewPaxos create a new paxos
func NewPaxos(paxosSequenceID int, numParticipant int, nodeIndex int, paxosRetry int) *Paxos {
	seqGen := newSeqGen(nodeIndex, numParticipant)

	return &Paxos{
		paxosSequenceID: paxosSequenceID,
		nodeIndex:       nodeIndex,
		numParticipant:  numParticipant,
		paxosRetry:      paxosRetry,

		idGenerator: seqGen,
		ID:          nodeIndex,

		proposedID:   -1,
		state:        stateNoProposal,
		count:        0,
		value:        nil,
		chanMajority: make(chan bool),

		latestPrepareID:     -1,
		latestAcceptedID:    -1,
		latestAcceptedValue: extramessage.Block{},

		learnerCount: 0,
		learnerData:  make(map[int]int),

		chanEnd: make(chan bool),
	}
}

func (p *Paxos) propose(g *gossip.Gossiper, block *extramessage.Block) {
	go func() {
		// log.Printf("%s Call Propose value %s", g.identifier, block.Filename)
		if p.value == nil {
			p.value = block
		}
		for {
			id := p.idGenerator.GetNext()
			p.proposedID = id
			p.state = stateAwaitPromise

			// log.Printf("%s Propose value %s", g.identifier, p.value.Filename)

			// Phase 1
			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosPrepare: &extramessage.PaxosPrepare{
					PaxosSeqID: p.paxosSequenceID,
					ID:         id,
				},
			})

			// log.Printf("%s Proposed value %s", g.identifier, p.value.Filename)

			// Create timer
			timer := time.NewTimer(time.Duration(p.paxosRetry) * time.Second)

			select {
			case <-timer.C:
				// log.Printf("%s timeout phase 1 - %d s", g.identifier, p.paxosRetry)
				continue
			case <-p.chanMajority:
				// log.Printf("%s majority phase 1", g.identifier)
				timer.Stop()
				// Move to next step
			case <-p.chanEnd:
				// log.Printf("%s end", g.identifier)
				timer.Stop()
				return
			}

			// Phase 2
			log.Printf("Enter phase 2")
			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosPropose: &extramessage.PaxosPropose{
					PaxosSeqID: p.paxosSequenceID,
					ID:         id,
					Value:      *p.value,
				},
			})

			// Create timer
			timer = time.NewTimer(time.Duration(p.paxosRetry) * time.Second)

			select {
			case <-timer.C:
				// log.Printf("%s timeout phase 2 - %d s", g.identifier, p.paxosRetry)
				continue
			case <-p.chanMajority:
			case <-p.chanEnd:
			}

			// log.Printf("%s consensus", g.identifier)
			timer.Stop()
			return
		}
	}()
}

func (p *Paxos) stop() {
	defer func() {
		recover()
	}()
	close(p.chanEnd)
}

func (p *Paxos) handle(g *gossip.Gossiper, msg *extramessage.ExtraMessage) *extramessage.Block {
	// log.Printf("MANAGE handle")

	// Upon promise
	if msg.PaxosPrepare != nil {
		// log.Printf("Handle prepare %d vs %d ID %d", msg.PaxosPrepare.PaxosSeqID, msg.PaxosPrepare.PaxosSeqID, msg.PaxosPrepare.ID)
		p.uponPaxosPrepare(g, msg.PaxosPrepare)
	} else if msg.PaxosPromise != nil {
		// log.Printf("Handle promise %d vs %d ID %d", msg.PaxosPromise.PaxosSeqID, msg.PaxosPromise.PaxosSeqID, msg.PaxosPromise.IDp)
		p.uponPaxosPromise(g, msg.PaxosPromise)
	} else if msg.PaxosPropose != nil {
		// log.Printf("Handle propose %d vs %d ID %d", msg.PaxosPropose.PaxosSeqID, msg.PaxosPropose.PaxosSeqID, msg.PaxosPropose.ID)
		p.uponPaxosPropose(g, msg.PaxosPropose)
	} else if msg.PaxosAccept != nil {
		// log.Printf("Handle accept %d vs %d ID %d", msg.PaxosAccept.PaxosSeqID, msg.PaxosAccept.PaxosSeqID, msg.PaxosAccept.ID)
		return p.uponPaxosAccept(g, msg.PaxosAccept)
	}
	return nil
}

// --- Phase 1 ---

func (p *Paxos) uponPaxosPrepare(g *gossip.Gossiper, msg *extramessage.PaxosPrepare) {
	if msg.PaxosSeqID != p.paxosSequenceID {
		// log.Printf("Discarded prepare %d vs %d", msg.PaxosSeqID, p.paxosSequenceID)
		return // Discard
	}

	// log.Printf("MANAGE prepare %d vs %d ID %d", msg.PaxosSeqID, p.paxosSequenceID, p.ID)
	if p.latestPrepareID < msg.ID && p.latestAcceptedID == -1 {
		// send response back to the sender
		// log.Printf("Accept prepare %d vs %d", msg.PaxosSeqID, p.paxosSequenceID)
		p.latestPrepareID = msg.ID
		if p.proposedID != msg.ID {
			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosPromise: &extramessage.PaxosPromise{
					PaxosSeqID: p.paxosSequenceID,
					IDp:        msg.ID,
					IDa:        -1,
					Value:      extramessage.Block{},
				},
			})
		}
	} else if p.proposedID != msg.ID {
		// Send previously accepted value
		g.AddExtraMessage(&extramessage.ExtraMessage{
			PaxosPromise: &extramessage.PaxosPromise{
				PaxosSeqID: p.paxosSequenceID,
				IDp:        msg.ID,
				IDa:        p.latestAcceptedID,
				Value:      p.latestAcceptedValue,
			},
		})
	}
}

func (p *Paxos) uponPaxosPromise(g *gossip.Gossiper, msg *extramessage.PaxosPromise) {
	if msg.PaxosSeqID != p.paxosSequenceID {
		// log.Printf("Discarded promise %d vs %d", msg.PaxosSeqID, p.paxosSequenceID)
		return // Discard
	}

	if msg.IDp == p.proposedID && p.state == stateAwaitPromise {
		p.count++
		if msg.Value.Metahash != nil {
			p.value = &msg.Value
		}

		if p.count > p.numParticipant/2+1 {
			// next phase
			p.count = 0
			p.state = stateAwaitAccept
			p.chanMajority <- true
		}
	}
}

// --- Phase 2 ---

func (p *Paxos) uponPaxosPropose(g *gossip.Gossiper, msg *extramessage.PaxosPropose) {
	if msg.PaxosSeqID != p.paxosSequenceID {
		// log.Printf("Discarded propose %d vs %d", msg.PaxosSeqID, p.paxosSequenceID)
		return // Discard
	}

	// log.Printf("Propose %d vs %d - %d", msg.PaxosSeqID, p.paxosSequenceID, len(msg.Value.PreviousHash))
	if msg.ID >= p.latestPrepareID {
		p.latestAcceptedID = msg.ID
		p.latestAcceptedValue = msg.Value

		if p.proposedID != msg.ID {
			// Send to all an accept response
			g.AddExtraMessage(&extramessage.ExtraMessage{
				PaxosAccept: &extramessage.PaxosAccept{
					PaxosSeqID: msg.PaxosSeqID,
					ID:         msg.ID,
					Value:      msg.Value,
				},
			})
		}
	}
}

func (p *Paxos) uponPaxosAccept(g *gossip.Gossiper, msg *extramessage.PaxosAccept) *extramessage.Block {
	if msg.PaxosSeqID != p.paxosSequenceID {
		// log.Printf("Discarded accept %d vs %d", msg.PaxosSeqID, p.paxosSequenceID)
		return nil // Discard
	}

	count, ok := p.learnerData[msg.ID]
	if !ok {
		count = 1
	} else {
		count++
	}
	p.learnerData[msg.ID] = count

	if count > p.numParticipant/2+1 {
		p.acceptedCount = 0
		if msg.ID == p.proposedID && p.state == stateAwaitAccept {
			p.state = stateConsensus
			p.chanMajority <- true
			close(p.chanEnd)
		}
		return &msg.Value
	}
	return nil
}
