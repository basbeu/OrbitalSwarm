package extramessage

// Paxos messages. Feel free to move that in a separate file and/or package.

// ExtraMessage is carried by a rumor message.
type ExtraMessage struct {
	PaxosPrepare *PaxosPrepare
	PaxosPromise *PaxosPromise
	PaxosPropose *PaxosPropose
	PaxosAccept  *PaxosAccept
	PaxosTLC     *PaxosTLC
	SwarmInit    *SwarmInit
}

// Copy performs a deep copy of extra message
func (e *ExtraMessage) Copy() *ExtraMessage {
	var paxosPrepare *PaxosPrepare
	var paxosPromise *PaxosPromise
	var paxosPropose *PaxosPropose
	var paxosAccept *PaxosAccept
	var paxosTLC *PaxosTLC

	if e.PaxosPrepare != nil {
		paxosPrepare = new(PaxosPrepare)
		paxosPrepare.PaxosSeqID = e.PaxosPrepare.PaxosSeqID
		paxosPrepare.ID = e.PaxosPrepare.ID
	}

	if e.PaxosPromise != nil {
		paxosPromise = new(PaxosPromise)
		paxosPromise.PaxosSeqID = e.PaxosPromise.PaxosSeqID
		paxosPromise.IDp = e.PaxosPromise.IDp
		paxosPromise.IDa = e.PaxosPromise.IDa
		paxosPromise.Value = *(e.PaxosPromise.Value.Copy())
	}

	if e.PaxosPropose != nil {
		paxosPropose = new(PaxosPropose)
		paxosPropose.PaxosSeqID = e.PaxosPropose.PaxosSeqID
		paxosPropose.ID = e.PaxosPropose.ID
		paxosPropose.Value = *(e.PaxosPropose.Value.Copy())
	}

	if e.PaxosAccept != nil {
		paxosAccept = new(PaxosAccept)
		paxosAccept.PaxosSeqID = e.PaxosAccept.PaxosSeqID
		paxosAccept.ID = e.PaxosAccept.ID
		paxosAccept.Value = *(e.PaxosAccept.Value.Copy())
	}

	if e.PaxosTLC != nil {
		paxosTLC = new(PaxosTLC)
		paxosTLC.Block = *e.PaxosTLC.Block.Copy()
	}

	return &ExtraMessage{
		PaxosPrepare: paxosPrepare,
		PaxosPromise: paxosPromise,
		PaxosPropose: paxosPropose,
		PaxosAccept:  paxosAccept,
		PaxosTLC:     paxosTLC,
	}
}
