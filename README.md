# Multi-Paxos
Final Project for UW's CSE 550 Autumn 2021

## Running and Deploying Code
First clone the repository. 

### Running a Paxos Instance
To open each Paxos instance, please open a new terminal and run ```go run ./cmd/node/node.go --config cmd/config/config<i>.json``` in the root folder where ```<i>``` should be replaced with numeric value 1−4.


### Running a Client Instance 
To open a client instance, please open a new terminal and run ```go run ./cmd/client/client.go```

To issue a command from the client instance, in the corresponding terminal, type command. Available com-
mands are ```setid, lock <lockid> <ServerAddress>, unlock <lockid> <ServerAddress>```
