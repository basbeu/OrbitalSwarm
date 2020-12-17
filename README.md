# CS-438 Project

This is the reference implementation for CS438 Peerster system, homework 3.
 
p1 --> p2 --> p3 --> p1


Circular network. Initially, p1 has p2 as a known peer, p2 has p3 as a known peer, and p3 has p1 as a known peer.
p1 sends a message. The message should reach everyone and loop infinitely through the network, minus the cases when the very first transmission fails because of unreliable delivery over UDP. Additionally, every node adds the other two in their list of known neighbors.


`./peerster -UIPort=2222 -gossipAddr=127.0.0.1:5000 -name=p1 -peers=127.0.0.1:5001` 

`./peerster -UIPort=2223 -gossipAddr=127.0.0.1:5001 -name=p2 -peers=127.0.0.1:5002` 

`./peerster -UIPort=2224 -gossipAddr=127.0.0.1:5002 -name=p3 -peers=127.0.0.1:5000`

`./cli/cli -UIPort=2222`

# Run the integration test locally

The integration test checks that your gossiper can work with a reference implementation that we provide as a binary.

The test is written in `gossip/bingossip_test.go`. There, you will find the same test case as in `packets_test.go`, except that it uses a combination of multiple reference gossipers with your own gossiper implementation. This works by either calling the `initGossip` function, which uses the standard gossip factory, or the `initBinGossip`, which uses the `binfactory` that create a gossiper based on a provided binary.

The only thing you need to do to run the `bingossip_test.go` in local is to provide the correct binary for the binfactory to use. Right now, the factory is looking for `./hw0`, which means that the binary `hw0` is expected to be in the `gossip/` folder. Therefore, be sure that it is there (ie. you have `gossip/hw0`).

If your machine is a MacOs (Darwin) rather than Unix, there's the `gossip/hw0.osx` binary, which you first need to rename to `gossip/hw0` and then run the tests.


Depending on your platform, you may need a different binary that has been compiled for your platform. We provide binaries for MacOS and Linux (64 bits). So, be sure to use the correct version, rename the binary to hw0, and place it in the correct gossip/ folder.
