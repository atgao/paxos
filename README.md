# Multi-Paxos
Final Project for UW's CSE 550 Autumn 2021

## Group members
Alice Gao - atgao@cs.washington.edu

Kaiming Cheng - kaimingc@cs.washington.edu

Sirui Lu - siruilu@cs.washington.edu

## Environment
You need golang 1.17+ to run our Paxos implementation.

## How to run
### Paxos node (server)
To start a paxos node, please open a new terminal and run

```bash
go run ./cmd/node/node.go --config cmd/config/config<i>.json
```

Here the `i`s are for the four different nodes.

### Client
To start a client, please open a new terminal and run

```bash
go run ./cmd/client/client.go
```

This would start the client shell.
There are three available commands in the shell:

```bash
setid <clientId>
lock <lockId> <serverAddress>
# example
lock 1 localhost:1032 # The default server address in the config files
# You can get or change the address from the config file
unlock <lockId> <serverAddress>
```

The `setid` command will set the client id to clientId.
This should be executed before any other commands.

The `lock` command will try to acquire a specific lock.
If that lock is not available, and no deadlock is detected,
the command will block until the lock is available,
or there is a timeout.

If a deadlock is detected, the lock command will return immediately
with error message: deadlock detected.

The `unlock` command is similar.
