# NetPongTerm

NetPongTerm is a pong game that extends on multiple screen on multiple computer trough the network. You can watch a demo [here](https://youtu.be/lIkU9vaCdmQ).

## Usage

First, start the server:

```bash
netpongterm -mode=server
```

Then each clients from left to right. You can specify the server address with the `-server` parameter.

```bash
netpongterm
```

For the last client, add the `-last-client` option:

```bash
netpongterm -last-client
```

## Installation

To install this game, run:

```
go get github.com/moverest/netpongterm
go install github.com/moverest/netpongterm
```

You'll need the `go` compiler installed.
