Simple PBFT
------

This repository contains the golang code of simple pbft consensus implementation.

  
How to run
------

## Build

```shell script
go build 
```

### Start four pbft node

```shell script
./sr-bft pbft node -id 0
./sr-bft pbft node -id 1
./sr-bft pbft node -id 2
./sr-bft pbft node -id 3
```

### Start pbft client to send message

```shell script
./sr-bft pbft client
```


### Reference

- https://www.jianshu.com/p/78e2b3d3af62


### TODOS :
UPDATE CONFIG METHOD
FIX COMMUNICATION/NETWORKING -> PORT FOR CLIENTS / PORT FOR REPLICA CONSENSUS / PORT FOR STATE TRANSFER
ADD benchmarking
ADD checkpoints and state transfer https://arxiv.org/pdf/2110.04448.pdf
ADD client request batching
ADD service proxy 
ADD view Change
