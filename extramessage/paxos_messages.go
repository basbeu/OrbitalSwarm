package extramessage

import "go.dedis.ch/cs438/orbitalswarm/paxos/blk"

// PaxosPrepare describes a PREPARE request to an acceptor.
type PaxosPrepare struct {
	PaxosSeqID int
	ID         int
}

// PaxosPromise describes a PROMISE request made by an acceptor to a proposer.
// IDp is the ID the proposer sent. IDa is the highest ID the acceptor saw and
// Value is the value it commits to, if any.
// Value/ID
type PaxosPromise struct {
	PaxosSeqID int
	IDp        int

	IDa   int
	Value blk.Block
}

// PaxosPropose describes a PROPOSE request made by a proposer to an ACCEPTOR.
type PaxosPropose struct {
	PaxosSeqID int
	ID         int

	Value blk.Block
}

// PaxosAccept describes an ACCEPT request that is sent by an acceptor to its
// proposer and all the learners.
type PaxosAccept struct {
	PaxosSeqID int
	ID         int

	Value blk.Block
}

// PaxosTLC is the message sent by a node when it knows consensus has been reached
// for that block.
type PaxosTLC struct {
	Block blk.Block
}
