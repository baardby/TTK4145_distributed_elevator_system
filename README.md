TTK4145 Distributed Elevator System
===================================
This project was an extensive labproject where we designed and implemented a fault tolerant distributed elevator system containing N elevators with M floors. Each elevator is its own computer, which can communicate with the computers of other elevators via networking. The project was a part of the course TTK4145 Real-Time Programming, and involved implementing the system on three phsyical elevators. That is also the reason our code implementation uses N_ELEVATORS=3.

To achieve fault tolerance in a real physical elevator which will service people, our solution prioritized availability and partition tolerance over consistency.


Network Topology
----------------
Our solution to the project is based on a peer-to-peer network topology using periodic UDP broadcasts. Each elevator broadcasts their own view of the global queue of requests, as well as their own state. This lets the elevators distribute requests between each other, and in case one fails, redistribute the orders of the peer that has failed. Only the elevator which sent the message, can tell the others about its state, but elevator 1 can tell elevator 3 what requests elevator 2 will do. This introduces the problem of merging each elevator's view of all requests. Our solution to this was that each request is its own cyclic state machine, which was designed to make the elevators coordinate if requests are unconfirmed, confirmed, completed or non-active.

The network topology of the system does also have some influence from a master-slave topology. The elevator that has received the request button press, is the one that distributes and assigns the request. If one elevator stops receiving from one of its peers, it will assume that that peer has failed, and the elevator with the lowest ID that is still alive will redistribute the requests that was assigned to the failed peer. To prevent problems where elevators doesn't agree which elevator is assigned to a request, the elevator with the lowest ID will decide.

Elevators will also be able to detect if they themselves are faulty or have been obstructed for too long, and can therefore tell the others that it is faulty and can't complete any of its orders. It will still take part in the network in case the problem disappears and it can resume taking requests.


Inside One Elevator
-------------------
Our distributed elevator utilizes the Go programming language's capabilities for communicating goroutines to achieve concurrency within each elevator. Each elevator has nine goroutines:
- Floor sensor poller
- Stop button poller
- Obstruction sensor poller
- Call button poller
- Elevator controller
- Network listener
- Network sender
- Supervisor
- Global state manager (GSM)

The polling goroutines are used to achieve availability in the physical interface with the elevator. All button presses and sensor interactions are to be registered and taken care of by the elevator controller and GSM goroutine. The GSM goroutine handles the elevator's view of the global state, i.e. the state of the other elevators it has received and the global request queue. It uses inputs from the call button poller to add new requests and ensure correct transition in the state of the requests. The elevator controller takes the requests assigned to the elevator, tries to complete them in a sensible order and keeps track of the state of the physical elevator. This information is used by the supervisor goroutine to keep track of the health of the system. It will keep track of how long since the elevator has heard from its peers, how long the elevator has been obstructed, and how long it has been moving between floors, to determine if the others or itself has experienced failure. At last is the network listener which receives information from the network and notifies supervisor of who it received from, and GSM of what was received.


The Simulator
-------------
The details of how to run the simulator is explained here: (https://github.com/TTK4145/Simulator-v2). Make sure that the port used in main to connect to the elevator is the same used for the SimElevatorServer.
