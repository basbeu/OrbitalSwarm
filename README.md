# CS-438 OrbitalSwarm

Orbital Swarm has been developped in the context of course CS-438 given at EPFL.
It provides a distributed drone simulation.

To build and run the project. You need to have golang installed.
The project has been developed and tested on go1.14.12 linux/amd64.

Commands to build and run the project :
 ```
go build
./orbitalswarm
 ```
To run the tests without evaluation (evaluation tests take lot of times)  :
 ```
go test ./... -v -p 1 -short  
 ```
To run all the tests :
 ```
go test ./... -v -p 1
 ```
