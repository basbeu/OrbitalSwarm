package drone

// TO TEST PAXOS with naming

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"go.dedis.ch/cs438/orbitalswarm/pathgenerator"

	"go.dedis.ch/cs438/orbitalswarm/drone/mapping"
	"go.dedis.ch/cs438/orbitalswarm/paxos/blk"
	"gonum.org/v1/gonum/spatial/r3"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/orbitalswarm/extramessage"
	"go.dedis.ch/cs438/orbitalswarm/gossip"
	"go.dedis.ch/cs438/orbitalswarm/paxos"
)

var factory = gossip.GetFactory()

// The subfolders of the temporary folder that stores files
const (
	sharedDataFolder = "shared"
	downloadFolder   = "download"
	chunkSize        = 8192
)

// Test 1
// Node A proposes a file, we check that the other nodes got the PREPARE message
// Node A got the PROMISE messages back.
func TestGossiper_No_Contention_Single_Propose(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	inA := nA.getIns(ctx)
	inB := nB.getIns(ctx)
	inC := nC.getIns(ctx)
	inD := nD.getIns(ctx)
	inE := nE.getIns(ctx)

	metahash := hex.EncodeToString([]byte{0xAA})
	nA.drone.ProposeName(metahash, "test1.txt")
	time.Sleep(time.Second * 1)

	getPrepare := func(h *history) int {
		h.Lock()
		defer h.Unlock()

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosPrepare != nil {

				return 1
			}
		}

		return 0
	}

	getPromise := func(h *history) []*extramessage.PaxosPromise {
		res := []*extramessage.PaxosPromise{}

		h.Lock()
		defer h.Unlock()

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosPromise != nil {

				res = append(res, msg.message.Rumor.Extra.PaxosPromise)
			}
		}

		return res
	}

	// A should receive at least 3 Promise messages
	promises := getPromise(inA)
	require.True(t, len(promises) >= 3)

	// At least 3 nodes should receive the Prepare message
	bReceived := getPrepare(inB)
	cReceived := getPrepare(inC)
	dReceived := getPrepare(inD)
	eReceived := getPrepare(inE)

	require.True(t, bReceived+cReceived+dReceived+eReceived >= 3)
}

// Test 2
// Node A proposes a file, we check that the other nodes got the PROPOSE message
// Node receive the ACCEPT messages back.
func TestGossiper_No_Contention_Single_Accept(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	inA := nA.getIns(ctx)
	inB := nB.getIns(ctx)
	inC := nC.getIns(ctx)
	inD := nD.getIns(ctx)
	inE := nE.getIns(ctx)

	metahash := hex.EncodeToString([]byte{0xAA})
	nA.drone.ProposeName(metahash, "test1.txt")

	time.Sleep(time.Second * 1)

	getAccept := func(h *history) []*extramessage.PaxosAccept {
		res := []*extramessage.PaxosAccept{}

		h.Lock()
		defer h.Unlock()

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosAccept != nil {

				res = append(res, msg.message.Rumor.Extra.PaxosAccept)
			}
		}

		return res
	}

	getPropose := func(h *history) int {
		h.Lock()
		defer h.Unlock()

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosPropose != nil {

				return 1
			}
		}

		return 0
	}

	// A should receive at least 3 Accept messages
	accepts := getAccept(inA)
	require.True(t, len(accepts) >= 3)

	// At least 3 nodes should receive the Propose message
	bReceived := getPropose(inB)
	cReceived := getPropose(inC)
	dReceived := getPropose(inD)
	eReceived := getPropose(inE)

	require.True(t, bReceived+cReceived+dReceived+eReceived >= 3)
}

// Test 3
// Node A proposes a file, we check that at least 3 nodes should receive at
// least 3 ACCEPT messages.
func TestGossiper_No_Contention_Single_Consensus_Completion(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	inA := nA.getIns(ctx)
	inB := nB.getIns(ctx)
	inC := nC.getIns(ctx)
	inD := nD.getIns(ctx)
	inE := nE.getIns(ctx)

	metahash := hex.EncodeToString([]byte{0xAA})
	nA.drone.ProposeName(metahash, "test1.txt")

	time.Sleep(time.Second * 2)

	getAccept := func(h *history) int {
		res := []*extramessage.PaxosAccept{}

		h.Lock()
		defer h.Unlock()

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosAccept != nil {

				res = append(res, msg.message.Rumor.Extra.PaxosAccept)
			}
		}

		if len(res) >= 3 {
			return 1
		}

		return 0
	}

	aReceived := getAccept(inA)
	bReceived := getAccept(inB)
	cReceived := getAccept(inC)
	dReceived := getAccept(inD)
	eReceived := getAccept(inE)

	require.True(t, aReceived+bReceived+cReceived+dReceived+eReceived >= 3)
}

// Test 4
// Node A proposes a file, node B, C, and D are down. We check that Node A
// retries after the "paxos retry" timeout.
func TestGossiper_No_Contention_Single_Retry(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithPaxosRetry(2), WithAntiEntropy(1))
	defer nA.stop()

	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	nB.stop()
	nC.stop()
	nD.stop()
	nE.stop()

	outA := nA.getOuts(ctx)

	metahash := hex.EncodeToString([]byte{0xAA})
	go nA.drone.ProposeName(metahash, "test1.txt")

	time.Sleep(time.Second * 1)

	getPrepare := func(h *history) []*extramessage.PaxosPrepare {
		res := []*extramessage.PaxosPrepare{}

		h.Lock()
		defer h.Unlock()

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosPrepare != nil {

				res = append(res, msg.message.Rumor.Extra.PaxosPrepare)
			}
		}

		return res
	}

	aSent := getPrepare(outA)

	// Node A, as a proposer, should've gossiped at least one propose message
	require.True(t, len(aSent) >= 1)
	// all the PROPOSE message should have index 0
	for _, p := range aSent {
		require.True(t, p.ID == 0)
	}

	time.Sleep(time.Second * 3)

	// we expect that we find a PROPOSE message with ID = 0 + numNodes
	aSent = getPrepare(outA)
	ok := false
	for _, p := range aSent {
		if p.ID == 5 {
			ok = true
		}
	}

	require.True(t, ok)
}

// Test 8
// Node A proposes a name-metahash, we check that at least 3 nodes have the
// name-metahash as the first block or their chain.
func TestGossiper_No_Contention_Block_Consensus(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	content := []byte{0xAA}

	// Compute the metahash
	h := sha256.New()
	h.Write(content)
	chunk := h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash := h.Sum(nil)

	metahash := hex.EncodeToString(metaHash)
	nA.drone.ProposeName(metahash, "test1.txt")

	time.Sleep(time.Second * 2)

	aLast, aChain := nA.drone.naming.GetBlocks()
	bLast, bChain := nB.drone.naming.GetBlocks()
	cLast, cChain := nC.drone.naming.GetBlocks()
	dLast, dChain := nD.drone.naming.GetBlocks()
	eLast, eChain := nE.drone.naming.GetBlocks()

	checkBlock := func(last string, chain map[string]*blk.BlockContainer) int {
		if len(chain) == 1 {
			blockContainer := chain[aLast]
			block := blockContainer.Block
			expected := &blk.NamingBlock{
				BlockNum: 0,
				PrevHash: make([]byte, 32),
				Content: &blk.NamingBlockContent{
					Filename: "test1.txt",
					Metahash: metaHash,
				},
			}
			require.Equal(t, expected, block)

			return 1
		}

		return 0
	}

	aCheck := checkBlock(aLast, aChain)
	bCheck := checkBlock(bLast, bChain)
	cCheck := checkBlock(cLast, cChain)
	dCheck := checkBlock(dLast, dChain)
	eCheck := checkBlock(eLast, eChain)

	require.True(t, aCheck+bCheck+cCheck+dCheck+eCheck >= 3)
}

// Test 9
// Node A proposes a name-metahash, we check that at least 3 nodes received at
// least 3 TLC messages.
func TestGossiper_No_Contention_Block_TLC_Consensus(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inA := nA.getIns(ctx)
	inB := nB.getIns(ctx)
	inC := nC.getIns(ctx)
	inD := nD.getIns(ctx)
	inE := nE.getIns(ctx)

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	content := []byte{0xAA}

	// Compute the metahash
	h := sha256.New()
	h.Write(content)
	chunk := h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash := h.Sum(nil)

	metahash := hex.EncodeToString(metaHash)
	nA.drone.ProposeName(metahash, "test1.txt")

	time.Sleep(time.Second * 2)

	expectedBlock := &blk.NamingBlock{
		BlockNum: 0,
		PrevHash: make([]byte, 32),
		Content: &blk.NamingBlockContent{
			Filename: "test1.txt",
			Metahash: metaHash,
		},
	}

	getTLC := func(h *history) int {
		h.Lock()
		defer h.Unlock()

		count := 0

		for _, msg := range h.p {
			if msg.message.Rumor != nil &&
				msg.message.Rumor.Extra != nil &&
				msg.message.Rumor.Extra.PaxosTLC != nil {

				require.Equal(t, expectedBlock, msg.message.Rumor.Extra.PaxosTLC.Value.Block)
				count++
			}
		}

		if count >= 3 {
			return 1
		}

		return 0
	}

	aTLC := getTLC(inA)
	bTLC := getTLC(inB)
	cTLC := getTLC(inC)
	dTLC := getTLC(inD)
	eTLC := getTLC(inE)

	require.True(t, aTLC+bTLC+cTLC+dTLC+eTLC >= 3)
}

// Test 10
// Node A and B sends a proposal at the same time. We check that at least 3
// nodes agree on the same proposal.
func TestGossiper_Contention_Single_Block_Consensus(t *testing.T) {
	name1 := "test1.txt"
	name2 := "test2.txt"

	content1 := []byte{0xAA}
	content2 := []byte{0xBB}

	var metaHash1 []byte
	var metaHash2 []byte

	// Compute the metahash
	h := sha256.New()
	h.Write(content1)
	chunk := h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash1 = h.Sum(nil)

	h = sha256.New()
	h.Write(content2)
	chunk = h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash2 = h.Sum(nil)

	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	metahash1 := hex.EncodeToString(metaHash1)
	metahash2 := hex.EncodeToString(metaHash2)

	go nA.drone.ProposeName(metahash1, name1)
	go nB.drone.ProposeName(metahash2, name2)

	time.Sleep(time.Second * 2)

	aLast, aChain := nA.drone.naming.GetBlocks()
	bLast, bChain := nB.drone.naming.GetBlocks()
	cLast, cChain := nC.drone.naming.GetBlocks()
	dLast, dChain := nD.drone.naming.GetBlocks()
	eLast, eChain := nE.drone.naming.GetBlocks()

	getFirst := func(last string, chain map[string]*blk.BlockContainer) (*blk.BlockContainer, bool) {
		if last == "" {
			return &blk.BlockContainer{}, false
		}

		for {
			blockContainer := chain[last]
			block := blockContainer.Block.(*blk.NamingBlock)
			if block.BlockNum == 0 {
				return blockContainer, true
			}

			last = hex.EncodeToString(block.PrevHash)
		}
	}

	mergeChain := func(block *blk.BlockContainer, chainCount map[string]int, chainMerged map[string]*blk.BlockContainer) {
		key := hex.EncodeToString(block.Hash())

		_, ok := chainCount[key]
		if !ok {
			chainCount[key] = 0
		}

		chainCount[key]++
		chainMerged[key] = block
	}

	countBlocks := map[string]int{}
	mergedBlocks := map[string]*blk.BlockContainer{}

	aFirst, ok := getFirst(aLast, aChain)
	if ok {
		mergeChain(aFirst, countBlocks, mergedBlocks)
	}

	bFirst, ok := getFirst(bLast, bChain)
	if ok {
		mergeChain(bFirst, countBlocks, mergedBlocks)
	}

	cFirst, ok := getFirst(cLast, cChain)
	if ok {
		mergeChain(cFirst, countBlocks, mergedBlocks)
	}

	dFirst, ok := getFirst(dLast, dChain)
	if ok {
		mergeChain(dFirst, countBlocks, mergedBlocks)
	}

	eFirst, ok := getFirst(eLast, eChain)
	if ok {
		mergeChain(eFirst, countBlocks, mergedBlocks)
	}

	// There should be at least one block with a count of 3

	threeFound := false
	keyFound := ""
	for k, v := range countBlocks {
		if v >= 3 {
			threeFound = true
			keyFound = k
		}
	}

	require.True(t, threeFound)

	// We check that the block with at least three occurrences is the expected
	// one.

	block1 := &blk.BlockContainer{
		Type: "NamingBlock",
		Block: &blk.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
			Content: &blk.NamingBlockContent{
				Filename: name1,
				Metahash: metaHash1,
			},
		},
	}

	block2 := &blk.BlockContainer{
		Type: "NamingBlock",
		Block: &blk.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
			Content: &blk.NamingBlockContent{
				Filename: name2,
				Metahash: metaHash2,
			},
		},
	}

	allPotentialBlocks := []*blk.BlockContainer{
		block1,
		block2,
	}

	require.True(t, threeFound)
	require.Contains(t, allPotentialBlocks, mergedBlocks[keyFound])
}

// Test 11
// Node A and B make a proposal. The proposal that is rejected for block 1
// should be accepted in block 2, with at least 3 nodes having that proposal.
func TestGossiper_Contention_TLC_Retry(t *testing.T) {
	name1 := "test1.txt"
	name2 := "test2.txt"

	content1 := []byte{0xAA}
	content2 := []byte{0xBB}

	var metaHash1 []byte
	var metaHash2 []byte

	// Compute the metahash
	h := sha256.New()
	h.Write(content1)
	chunk := h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash1 = h.Sum(nil)

	h = sha256.New()
	h.Write(content2)
	chunk = h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash2 = h.Sum(nil)

	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	metahash1 := hex.EncodeToString(metaHash1)
	metahash2 := hex.EncodeToString(metaHash2)

	go nA.drone.ProposeName(metahash1, name1)
	go nB.drone.ProposeName(metahash2, name2)

	time.Sleep(time.Second * 10)

	aLast, aChain := nA.drone.naming.GetBlocks()
	bLast, bChain := nB.drone.naming.GetBlocks()
	cLast, cChain := nC.drone.naming.GetBlocks()
	dLast, dChain := nD.drone.naming.GetBlocks()
	eLast, eChain := nE.drone.naming.GetBlocks()

	getBlockByIndex := func(index int, last string, chain map[string]*blk.BlockContainer) (*blk.BlockContainer, bool) {
		if last == "" {
			return &blk.BlockContainer{}, false
		}

		for {
			blockContainer := chain[last]
			block := blockContainer.Block.(*blk.NamingBlock)
			if block.BlockNum == index {
				return blockContainer, true
			}

			// no block found
			if block.BlockNum == 0 {
				return &blk.BlockContainer{}, false
			}

			last = hex.EncodeToString(block.PrevHash)
		}
	}

	mergeChain := func(block *blk.BlockContainer, chainCount map[string]int, chainMerged map[string]*blk.BlockContainer) {
		key := hex.EncodeToString(block.Hash())

		_, ok := chainCount[key]
		if !ok {
			chainCount[key] = 0
		}

		chainCount[key]++
		chainMerged[key] = block
	}

	countBlocks := map[string]int{}
	mergedBlocks := map[string]*blk.BlockContainer{}

	aFirst, ok := getBlockByIndex(0, aLast, aChain)
	if ok {
		mergeChain(aFirst, countBlocks, mergedBlocks)
	}

	bFirst, ok := getBlockByIndex(0, bLast, bChain)
	if ok {
		mergeChain(bFirst, countBlocks, mergedBlocks)
	}

	cFirst, ok := getBlockByIndex(0, cLast, cChain)
	if ok {
		mergeChain(cFirst, countBlocks, mergedBlocks)
	}

	dFirst, ok := getBlockByIndex(0, dLast, dChain)
	if ok {
		mergeChain(dFirst, countBlocks, mergedBlocks)
	}

	eFirst, ok := getBlockByIndex(0, eLast, eChain)
	if ok {
		mergeChain(eFirst, countBlocks, mergedBlocks)
	}

	// There should be at least one block with a count of 3 for block 0

	threeFound := false
	keyFound := ""
	for k, v := range countBlocks {
		if v >= 3 {
			threeFound = true
			keyFound = k
		}
	}

	require.True(t, threeFound)

	// We check that the block with at least three occurrences is the expected
	// one.

	block1 := &blk.BlockContainer{
		Type: "NamingBlock",
		Block: &blk.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
			Content: &blk.NamingBlockContent{
				Filename: name1,
				Metahash: metaHash1,
			},
		},
	}

	block2 := &blk.BlockContainer{
		Type: "NamingBlock",
		Block: &blk.NamingBlock{
			BlockNum: 0,
			PrevHash: make([]byte, 32),
			Content: &blk.NamingBlockContent{
				Filename: name2,
				Metahash: metaHash2,
			},
		},
	}

	allPotentialBlocks := []*blk.BlockContainer{
		block1,
		block2,
	}

	firstBlock := mergedBlocks[keyFound]

	require.True(t, threeFound)
	require.Contains(t, allPotentialBlocks, firstBlock)

	// We do the same for the second block

	countBlocks = map[string]int{}
	mergedBlocks = map[string]*blk.BlockContainer{}

	aFirst, ok = getBlockByIndex(1, aLast, aChain)
	if ok {
		mergeChain(aFirst, countBlocks, mergedBlocks)
	}

	bFirst, ok = getBlockByIndex(1, bLast, bChain)
	if ok {
		mergeChain(bFirst, countBlocks, mergedBlocks)
	}

	cFirst, ok = getBlockByIndex(1, cLast, cChain)
	if ok {
		mergeChain(cFirst, countBlocks, mergedBlocks)
	}

	dFirst, ok = getBlockByIndex(1, dLast, dChain)
	if ok {
		mergeChain(dFirst, countBlocks, mergedBlocks)
	}

	eFirst, ok = getBlockByIndex(1, eLast, eChain)
	if ok {
		mergeChain(eFirst, countBlocks, mergedBlocks)
	}

	// There should be at least one block with a count of 3 for block 0

	threeFound = false
	keyFound = ""
	for k, v := range countBlocks {
		if v >= 3 {
			threeFound = true
			keyFound = k
		}
	}

	require.True(t, threeFound)

	// We check that the block with at least three occurrences is not the same
	// one as the first block.

	require.NotEqual(t, mergedBlocks[keyFound].Block.GetContent().(*blk.NamingBlockContent).Metahash, firstBlock.Block.GetContent().(*blk.NamingBlockContent).Metahash)
	require.NotEqual(t, mergedBlocks[keyFound].Block.GetContent().(*blk.NamingBlockContent).Filename, firstBlock.Block.GetContent().(*blk.NamingBlockContent).Filename)

	require.True(t, threeFound)

	block1 = &blk.BlockContainer{
		Type: "NamingBlock",
		Block: &blk.NamingBlock{
			BlockNum: 1,
			PrevHash: firstBlock.Hash(),
			Content: &blk.NamingBlockContent{
				Filename: name1,
				Metahash: metaHash1,
			},
		},
	}

	block2 = &blk.BlockContainer{
		Type: "NamingBlock",
		Block: &blk.NamingBlock{
			BlockNum: 1,
			PrevHash: firstBlock.Hash(),
			Content: &blk.NamingBlockContent{
				Filename: name2,
				Metahash: metaHash2,
			},
		},
	}

	allPotentialBlocks = []*blk.BlockContainer{
		block1,
		block2,
	}

	require.Contains(t, allPotentialBlocks, mergedBlocks[keyFound])
}

// Test 12
// We make Node A send multiple proposals, which should all be included in
// succeeding blocks.
func TestGossiper_No_Contention_Long_Blockchain(t *testing.T) {
	numProposals := 6
	numNodes := 5
	proposalIndex := 0

	maxRetry := 10
	waitRetry := time.Second * 5

	time.Sleep(time.Second * 10)

	rand.Seed(time.Now().UnixNano())

	proposeBlock := func(node *nodeInfo) (string, []byte) {

		fileNameBuf := make([]byte, 8)
		_, err := rand.Read(fileNameBuf)
		require.NoError(t, err)

		fileName := hex.EncodeToString(fileNameBuf)

		content := make([]byte, 4)
		_, err = rand.Read(content)
		require.NoError(t, err)

		// Compute the metahash
		h := sha256.New()
		h.Write(content)
		chunk := h.Sum(nil)

		h = sha256.New()
		h.Write(chunk)
		metaHash := h.Sum(nil)

		_, blocks := node.drone.naming.GetBlocks()
		blockN := len(blocks)

		metahash := hex.EncodeToString(metaHash)
		node.drone.ProposeName(metahash, fileName)

		lastBlock, blocks := node.drone.naming.GetBlocks()

		for len(blocks) == blockN {
			time.Sleep(time.Millisecond * 500)
			lastBlock, blocks = node.drone.naming.GetBlocks()
		}

		blockContainer := blocks[lastBlock]
		block := blockContainer.Block.(*blk.NamingBlock)

		require.Equal(t, metaHash, block.Content.(*blk.NamingBlockContent).Metahash)
		require.Equal(t, fileName, block.Content.(*blk.NamingBlockContent).Filename)
		require.Equal(t, proposalIndex, block.BlockNum)

		proposalIndex++

		return fileName, metaHash
	}

	nodes := make([]*nodeInfo, numNodes)
	bagNodes := make(map[string]nodeInfo)

	for i := 0; i < numNodes; i++ {
		n := createAndStartNode(t, string(rune('A'+i)), WithNodeIndex(i), WithNumParticipants(numNodes), WithAntiEntropy(1))
		nodes[i] = &n
		bagNodes[string(rune('A'+i))] = n
	}

	defer stopNodes(nodes...)

	lb := newLinkBuilder(bagNodes)
	lb.connectAll()

	var filename string
	var metaHash []byte

	for i := 0; i < numProposals; i++ {
		filename, metaHash = proposeBlock(nodes[0])
	}

	// We check that at least N/2 + 1 other nodes have 6 blocks

	ok := false

	for i := 0; i < maxRetry; i++ {

		hasBlocks := 0

		checkBlocks := func(node *nodeInfo) {
			last, chain := node.drone.naming.GetBlocks()

			time.Sleep(time.Millisecond * 100)

			if len(chain) == numProposals {
				hasBlocks++

				lastBlockContainer := chain[last]
				lastBlock := lastBlockContainer.Block.(*blk.NamingBlock)
				require.Equal(t, filename, lastBlock.GetContent().(*blk.NamingBlockContent).Filename)
				require.Equal(t, metaHash, lastBlock.GetContent().(*blk.NamingBlockContent).Metahash)
				require.Equal(t, numProposals-1, lastBlock.BlockNum)
			}
		}

		time.Sleep(time.Second)

		for _, n := range nodes {
			checkBlocks(n)
		}

		if hasBlocks >= numNodes/2+1 {
			ok = true
			break
		}

		time.Sleep(waitRetry)
	}

	require.True(t, ok, "after %d retries, didn't find enough nodes "+
		"with the expected number of blocks", maxRetry)
}

// Test 13
// Same as test 12 but with more nodes
func TestGossiper_No_Contention_Higher_Nodes_Long_Blockchain(t *testing.T) {
	numProposals := 4
	numNodes := 11

	maxRetry := 10
	waitRetry := time.Second * 5

	proposalIndex := 0

	time.Sleep(time.Second * 10)

	rand.Seed(time.Now().UnixNano())

	proposeBlock := func(node *nodeInfo) (string, []byte) {

		fileNameBuf := make([]byte, 8)
		_, err := rand.Read(fileNameBuf)
		require.NoError(t, err)

		fileName := hex.EncodeToString(fileNameBuf)

		content := make([]byte, 4)
		_, err = rand.Read(content)
		require.NoError(t, err)

		// Compute the metahash
		h := sha256.New()
		h.Write(content)
		chunk := h.Sum(nil)

		h = sha256.New()
		h.Write(chunk)
		metaHash := h.Sum(nil)

		_, blocks := node.drone.naming.GetBlocks()
		blockN := len(blocks)

		metahash := hex.EncodeToString(metaHash)
		node.drone.ProposeName(metahash, fileName)

		lastBlock, blocks := node.drone.naming.GetBlocks()

		for len(blocks) == blockN {
			time.Sleep(time.Millisecond * 500)
			lastBlock, blocks = node.drone.naming.GetBlocks()
		}

		blockContainer := blocks[lastBlock]
		block := blockContainer.Block.(*blk.NamingBlock)

		require.Equal(t, metaHash, block.GetContent().(*blk.NamingBlockContent).Metahash)
		require.Equal(t, fileName, block.GetContent().(*blk.NamingBlockContent).Filename)
		require.Equal(t, proposalIndex, block.BlockNum)

		proposalIndex++

		return fileName, metaHash
	}

	nodes := make([]*nodeInfo, numNodes)
	bagNodes := make(map[string]nodeInfo)

	for i := 0; i < numNodes; i++ {
		n := createAndStartNode(t, string(rune('A'+i)), WithNodeIndex(i), WithNumParticipants(numNodes), WithAntiEntropy(1))
		nodes[i] = &n
		bagNodes[string(rune('A'+i))] = n
	}

	defer stopNodes(nodes...)

	lb := newLinkBuilder(bagNodes)
	lb.connectAll()

	var filename string
	var metaHash []byte

	for i := 0; i < numProposals; i++ {
		filename, metaHash = proposeBlock(nodes[0])
	}

	// We check that at least N/2 - 1 other nodes have 6 blocks

	ok := false

	for i := 0; i < maxRetry; i++ {
		hasBlocks := 0

		checkBlocks := func(node *nodeInfo) {
			last, chain := node.drone.naming.GetBlocks()

			if len(chain) == numProposals {
				hasBlocks++

				lastBlockContainer := chain[last]
				lastBlock := lastBlockContainer.Block.(*blk.NamingBlock)
				require.Equal(t, filename, lastBlock.GetContent().(*blk.NamingBlockContent).Filename)
				require.Equal(t, metaHash, lastBlock.GetContent().(*blk.NamingBlockContent).Metahash)
				require.Equal(t, numProposals-1, lastBlock.BlockNum)
			}
		}

		for _, n := range nodes {
			checkBlocks(n)
		}

		if hasBlocks >= numNodes/2+1 {
			ok = true
			break
		}
		time.Sleep(waitRetry)
	}

	require.True(t, ok, "after %d retries, didn't find enough nodes "+
		"with the expected number of blocks", maxRetry)
}

// Test 14
// Multiple nodes try to send a proposal. We check that at least 3 nodes have
// the same proposal in their last and only block
func TestGossiper_Contention_Long_Blockchain(t *testing.T) {
	numNodes := 5
	numProposals := 4

	maxRetry := 30
	waitRetry := time.Second * 10

	time.Sleep(time.Second * 10)

	rand.Seed(time.Now().UnixNano())

	proposeBlock := func(node *nodeInfo) (string, []byte) {

		fileNameBuf := make([]byte, 8)
		_, err := rand.Read(fileNameBuf)
		require.NoError(t, err)

		fileName := hex.EncodeToString(fileNameBuf)

		content := make([]byte, 4)
		_, err = rand.Read(content)
		require.NoError(t, err)

		// Compute the metahash
		h := sha256.New()
		h.Write(content)
		chunk := h.Sum(nil)

		h = sha256.New()
		h.Write(chunk)
		metaHash := h.Sum(nil)

		metahash := hex.EncodeToString(metaHash)
		node.drone.ProposeName(metahash, fileName)

		return fileName, metaHash
	}

	nodes := make([]*nodeInfo, numNodes)
	bagNodes := make(map[string]nodeInfo)

	for i := 0; i < numNodes; i++ {
		n := createAndStartNode(t, string(rune('A'+i)), WithNodeIndex(i), WithNumParticipants(numNodes), WithAntiEntropy(1))
		nodes[i] = &n
		bagNodes[string(rune('A'+i))] = n
	}

	defer stopNodes(nodes...)

	lb := newLinkBuilder(bagNodes)
	lb.connectAll()

	fileNames := make([]string, numProposals)
	metaHashes := make([][]byte, numProposals)

	for i := 0; i < numProposals; i++ {
		go func(i int) {
			filename, metahash := proposeBlock(nodes[i])
			fileNames[i] = filename
			metaHashes[i] = metahash
		}(i)
	}

	ok := false

	time.Sleep(waitRetry)

	for i := 0; i < maxRetry; i++ {

		numFound := 0

		// We check that at least N/2 - 1 other nodes have numProposals blocks

		for _, n := range nodes {
			_, chain := n.drone.naming.GetBlocks()

			time.Sleep(time.Millisecond * 100)

			if len(chain) == numProposals {
				numFound++
			}
		}

		if numFound >= numNodes/2+1 {
			ok = true
			break
		}

		time.Sleep(waitRetry)
	}

	require.True(t, ok, "consensus not reached after %d retries", maxRetry)
}

// Test 15
// Same as test 14 but with more nodes
func TestGossiper_Contention_Higher_Nodes_Long_Blockchain(t *testing.T) {
	numNodes := 11
	numProposals := 4

	maxRetry := 30
	waitRetry := time.Second * 10

	time.Sleep(time.Second * 10)

	rand.Seed(time.Now().UnixNano())

	proposeBlock := func(node *nodeInfo) (string, []byte) {

		fileNameBuf := make([]byte, 8)
		_, err := rand.Read(fileNameBuf)
		require.NoError(t, err)

		fileName := hex.EncodeToString(fileNameBuf)

		content := make([]byte, 4)
		_, err = rand.Read(content)
		require.NoError(t, err)

		// Compute the metahash
		h := sha256.New()
		h.Write(content)
		chunk := h.Sum(nil)

		h = sha256.New()
		h.Write(chunk)
		metaHash := h.Sum(nil)

		metahash := hex.EncodeToString(metaHash)
		node.drone.ProposeName(metahash, fileName)

		return fileName, metaHash
	}

	nodes := make([]*nodeInfo, numNodes)
	bagNodes := make(map[string]nodeInfo)

	for i := 0; i < numNodes; i++ {
		n := createAndStartNode(t, string(rune('A'+i)), WithNodeIndex(i), WithNumParticipants(numNodes), WithAntiEntropy(1))
		nodes[i] = &n
		bagNodes[string(rune('A'+i))] = n
	}

	defer stopNodes(nodes...)

	lb := newLinkBuilder(bagNodes)
	lb.connectAll()

	fileNames := make([]string, numProposals)
	metaHashes := make([][]byte, numProposals)

	for i := 0; i < numProposals; i++ {
		go func(i int) {
			filename, metahash := proposeBlock(nodes[i])
			fileNames[i] = filename
			metaHashes[i] = metahash
		}(i)
	}

	ok := false

	time.Sleep(waitRetry)

	for i := 0; i < maxRetry; i++ {

		numFound := 0

		// We check that at least N/2 - 1 other nodes have numProposals blocks

		for _, n := range nodes {
			_, chain := n.drone.naming.GetBlocks()

			time.Sleep(time.Millisecond * 100)

			if len(chain) == numProposals {
				numFound++
			}
		}

		if numFound >= numNodes/2+1 {
			ok = true
			break
		}

		time.Sleep(waitRetry)
	}

	require.True(t, ok, "consensus not reached after %d retries", maxRetry)
}

// Test 16
// We check the uniqueness of a filename on the chain. If a filename has already
// been recorded on the chain then it should not be able to store the same
// filename again.
func TestGossiper_Unique_Filename(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	// D register a filename

	filename := "test1.txt"
	content := []byte{0xAA}

	// Compute the metahash
	h := sha256.New()
	h.Write(content)
	chunk := h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash := h.Sum(nil)

	time.Sleep(time.Second)

	metahash := hex.EncodeToString(metaHash)
	nD.drone.ProposeName(metahash, filename)

	time.Sleep(time.Second * 10)

	aLast, aChain := nA.drone.naming.GetBlocks()
	bLast, bChain := nB.drone.naming.GetBlocks()
	cLast, cChain := nC.drone.naming.GetBlocks()
	dLast, dChain := nD.drone.naming.GetBlocks()
	eLast, eChain := nE.drone.naming.GetBlocks()

	checkBlock := func(last string, chain map[string]*blk.BlockContainer) int {
		if len(chain) == 1 {
			block := chain[aLast]
			expected := &blk.BlockContainer{
				Type: "NamingBlock",
				Block: &blk.NamingBlock{
					BlockNum: 0,
					PrevHash: make([]byte, 32),
					Content: &blk.NamingBlockContent{
						Filename: filename,
						Metahash: metaHash,
					},
				},
			}
			require.Equal(t, expected, block)

			return 1
		}

		return 0
	}

	aCheck := checkBlock(aLast, aChain)
	bCheck := checkBlock(bLast, bChain)
	cCheck := checkBlock(cLast, cChain)
	dCheck := checkBlock(dLast, dChain)
	eCheck := checkBlock(eLast, eChain)

	require.True(t, aCheck+bCheck+cCheck+dCheck+eCheck >= 3)

	// C tries to register the same filename

	content = []byte{0xBB}
	// Compute the metahash
	h = sha256.New()
	h.Write(content)
	chunk = h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash = h.Sum(nil)

	metahash = hex.EncodeToString(metaHash)
	nC.drone.ProposeName(metahash, filename)

	time.Sleep(time.Second * 2)

	_, aChain = nA.drone.naming.GetBlocks()
	_, bChain = nB.drone.naming.GetBlocks()
	_, cChain = nC.drone.naming.GetBlocks()
	_, dChain = nD.drone.naming.GetBlocks()
	_, eChain = nE.drone.naming.GetBlocks()

	require.Len(t, aChain, 1)
	require.Len(t, bChain, 1)
	require.Len(t, cChain, 1)
	require.Len(t, dChain, 1)
	require.Len(t, eChain, 1)
}

// Test 17
// We check the uniqueness of a metahash on the chain. If a metahash has already
// been recorded on the chain then it should not be able to store the same
// filename again.
func TestGossiper_Unique_Metahash(t *testing.T) {
	nA := createAndStartNode(t, "NA", WithNodeIndex(0), WithNumParticipants(5), WithAntiEntropy(1))
	nB := createAndStartNode(t, "NB", WithNodeIndex(1), WithNumParticipants(5), WithAntiEntropy(1))
	nC := createAndStartNode(t, "NC", WithNodeIndex(2), WithNumParticipants(5), WithAntiEntropy(1))
	nD := createAndStartNode(t, "ND", WithNodeIndex(3), WithNumParticipants(5), WithAntiEntropy(1))
	nE := createAndStartNode(t, "NE", WithNodeIndex(4), WithNumParticipants(5), WithAntiEntropy(1))

	defer stopNodes(&nA, &nB, &nC, &nD, &nE)

	nodes := map[string]nodeInfo{
		"A": nA, "B": nB, "C": nC, "D": nD, "E": nE,
	}

	lb := newLinkBuilder(nodes)
	lb.connectAll()

	filename := "test1.txt"
	content := []byte{0xAA}

	// Compute the metahash
	h := sha256.New()
	h.Write(content)
	chunk := h.Sum(nil)

	h = sha256.New()
	h.Write(chunk)
	metaHash := h.Sum(nil)

	metahash := hex.EncodeToString(metaHash)
	nC.drone.ProposeName(metahash, filename)

	time.Sleep(time.Second * 2)

	aLast, aChain := nA.drone.naming.GetBlocks()
	bLast, bChain := nB.drone.naming.GetBlocks()
	cLast, cChain := nC.drone.naming.GetBlocks()
	dLast, dChain := nD.drone.naming.GetBlocks()
	eLast, eChain := nE.drone.naming.GetBlocks()

	checkBlock := func(last string, chain map[string]*blk.BlockContainer) int {
		if len(chain) == 1 {
			block := chain[aLast]
			expected := &blk.BlockContainer{
				Type: "NamingBlock",
				Block: &blk.NamingBlock{
					BlockNum: 0,
					PrevHash: make([]byte, 32),
					Content: &blk.NamingBlockContent{
						Filename: filename,
						Metahash: metaHash,
					},
				},
			}
			require.Equal(t, expected, block)

			return 1
		}

		return 0
	}

	aCheck := checkBlock(aLast, aChain)
	bCheck := checkBlock(bLast, bChain)
	cCheck := checkBlock(cLast, cChain)
	dCheck := checkBlock(dLast, dChain)
	eCheck := checkBlock(eLast, eChain)

	require.True(t, aCheck+bCheck+cCheck+dCheck+eCheck >= 3)

	// Try to register the same content with a different filename

	filename = "test2.txt"

	nC.drone.ProposeName(metahash, filename)

	time.Sleep(time.Second * 2)

	_, aChain = nA.drone.naming.GetBlocks()
	_, bChain = nB.drone.naming.GetBlocks()
	_, cChain = nC.drone.naming.GetBlocks()
	_, dChain = nD.drone.naming.GetBlocks()
	_, eChain = nE.drone.naming.GetBlocks()

	require.Len(t, aChain, 1)
	require.Len(t, bChain, 1)
	require.Len(t, cChain, 1)
	require.Len(t, dChain, 1)
	require.Len(t, eChain, 1)
}

// -----------------------------------------------------------------------------
// Utility functions

type nodeInfo struct {
	addr     string
	id       string
	gossiper gossip.BaseGossiper
	stream   chan gossip.GossipPacket // optional

	template nodeTemplate
	drone    *Drone
}

func newNodeInfo(g gossip.BaseGossiper, addr string, t nodeTemplate) *nodeInfo {
	return &nodeInfo{
		id:       g.GetIdentifier(),
		addr:     addr,
		gossiper: g,
		template: t,
	}
}

func (n nodeInfo) getIns(ctx context.Context) *history {
	watchIn := &history{p: make([]packet, 0)}
	chanIn := n.gossiper.Watch(ctx, true)

	go func() {
		for p := range chanIn {
			watchIn.Lock()
			watchIn.p = append(watchIn.p, packet{p.Addr, p.Msg})
			// fmt.Println(p)
			watchIn.Unlock()
		}
	}()

	return watchIn
}

func (n nodeInfo) getOuts(ctx context.Context) *history {
	watchOut := &history{p: make([]packet, 0)}
	chanOut := n.gossiper.Watch(ctx, false)

	go func() {
		for p := range chanOut {
			watchOut.Lock()
			watchOut.p = append(watchOut.p, packet{p.Addr, p.Msg})
			// fmt.Println(p)
			watchOut.Unlock()
		}
	}()

	return watchOut
}

func (n nodeInfo) getCallbacks(ctx context.Context) *history {
	msgs := &history{p: make([]packet, 0)}

	n.gossiper.RegisterCallback(func(origin string, message gossip.GossipPacket) {
		msgs.Lock()
		msgs.p = append(msgs.p, packet{origin, message})
		msgs.Unlock()
	})

	return msgs
}

func (n *nodeInfo) stop() {
	n.gossiper.Stop()
}

func createAndStartNode(t *testing.T, name string, opts ...nodeOption) nodeInfo {

	n := createNode(t, factory, "127.0.0.1:0", name, opts...)

	startNodesBlocking(t, n.gossiper)
	go n.drone.Run()

	n.addr = n.gossiper.GetLocalAddr()

	return n
}

// createNode creates a node but don't start it. The address of the node is
// unknown until it has been started.
func createNode(t *testing.T, fac gossip.GossipFactory, addr, name string, opts ...nodeOption) nodeInfo {

	template := newNodeTemplate().apply(opts...)

	fullName := fmt.Sprintf("%v---%v", name, t.Name())

	var err error

	node, err := fac.New(addr, fullName, template.antiEntropy, template.routeTimer,
		template.numParticipants)
	require.NoError(t, err)

	require.Len(t, node.GetNodes(), 0)
	require.Equal(t, fullName, node.GetIdentifier())

	naming := paxos.NewNaming(template.numParticipants, template.nodeIndex, template.paxosRetry)
	drone := NewDrone(0, node.GetIdentifier(), addr, addr, node, node.GetNodes(), r3.Vec{}, mapping.NewHungarianMapper(), nil, naming, pathgenerator.NewGeneticPathGenerator())

	return nodeInfo{
		id:       node.GetIdentifier(),
		gossiper: node,
		template: *template,
		drone:    drone,
	}
}

type nodeTemplate struct {
	antiEntropy int
	routeTimer  int

	numParticipants int
	nodeIndex       int
	paxosRetry      int
}

func newNodeTemplate() *nodeTemplate {
	return &nodeTemplate{
		antiEntropy: 1000,
		routeTimer:  0,

		nodeIndex:       0,
		numParticipants: 1,
		paxosRetry:      1000,
	}
}

func (t *nodeTemplate) apply(opts ...nodeOption) *nodeTemplate {
	for _, opt := range opts {
		opt(t)
	}

	return t
}

type nodeOption func(*nodeTemplate)

func WithAntiEntropy(value int) nodeOption {
	return func(n *nodeTemplate) {
		n.antiEntropy = value
	}
}

func WithRouteTimer(value int) nodeOption {
	return func(n *nodeTemplate) {
		n.routeTimer = value
	}
}

func WithNumParticipants(num int) nodeOption {
	return func(n *nodeTemplate) {
		n.numParticipants = num
	}
}

func WithNodeIndex(index int) nodeOption {
	return func(n *nodeTemplate) {
		n.nodeIndex = index
	}
}

func WithPaxosRetry(d int) nodeOption {
	return func(n *nodeTemplate) {
		n.paxosRetry = d
	}
}

// startNodesBlocking waits until the node is started
func startNodesBlocking(t *testing.T, nodes ...gossip.BaseGossiper) {
	wg := new(sync.WaitGroup)
	wg.Add(len(nodes))
	for idx := range nodes {
		go func(i int) {
			defer wg.Done()
			ready := make(chan struct{})
			go nodes[i].Run(ready)
			<-ready
		}(idx)
	}
	wg.Wait()
}

type packet struct {
	origin  string
	message gossip.GossipPacket
}

type history struct {
	sync.Mutex
	p []packet
}

func (h *history) getPs() []packet {
	h.Lock()
	defer h.Unlock()

	return append([]packet{}, h.p...)
}

type linkBuilder struct {
	nodes map[string]nodeInfo
}

func newLinkBuilder(nodes map[string]nodeInfo) linkBuilder {
	return linkBuilder{nodes}
}

func (b linkBuilder) add(link string) {
	// we accept the following links:
	// "A --> B"
	// "A <-- B"
	// "A <-> B"

	parts := strings.Split(link, " ")
	nA := b.nodes[parts[0]]
	nB := b.nodes[parts[2]]

	switch parts[1] {
	case "-->":
		nA.gossiper.AddRoute(nB.id, nB.addr)
		nA.gossiper.AddAddresses(nB.addr)
	case "<--":
		nB.gossiper.AddRoute(nA.id, nA.addr)
		nB.gossiper.AddAddresses(nA.addr)
	case "<->":
		nA.gossiper.AddRoute(nB.id, nB.addr)
		nB.gossiper.AddRoute(nA.id, nA.addr)

		nA.gossiper.AddAddresses(nB.addr)
		nB.gossiper.AddAddresses(nA.addr)
	default:
		fmt.Println("[WARNING] unknown link", parts[1])
	}
}

func (b linkBuilder) connectAll() {
	for idFrom := range b.nodes {
		for idTo := range b.nodes {
			if idTo == idFrom {
				continue
			}

			b.add(idFrom + " --> " + idTo)
		}
	}
}

// stopNodes stops a set of nodes concurrently
func stopNodes(nodes ...*nodeInfo) {
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))

	for _, node := range nodes {
		go func(n *nodeInfo) {
			n.stop()
			wg.Done()
		}(node)
	}

	wg.Wait()
}
