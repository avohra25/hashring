# hashring

A small, thread-safe **consistent hashing ring** in Go with virtual nodes and
replica lookup — the technique used by Dynamo, Cassandra, and memcached clients
to spread keys across servers so that adding/removing a node only remaps ~1/N of keys.

## Features
- Virtual nodes (configurable replicas per node) for even load balance
- `Get(key)` – owning node via binary search over the sorted ring (O(log n))
- `GetN(key, n)` – n distinct successor nodes, for replication
- `Add` / `Remove` nodes at runtime; safe for concurrent use (`sync.RWMutex`)
- `ringsim` CLI demonstrating distribution and minimal remapping

## Usage
```go
r := hashring.New(160)
r.Add("cache-a", "cache-b", "cache-c")
node, _ := r.Get("user:42")
replicas, _ := r.GetN("user:42", 2)
```

## Run
Requires Go 1.21+.
```sh
go test -race ./...          # tests
go test -bench . ./...       # benchmark
go run ./cmd/ringsim -nodes 5 -keys 100000 -vnodes 160
```
Sample output:
```
Distribution of 100000 keys over 5 nodes (160 vnodes each):
  node-0    19037 (19.0%)
  ...
Adding 1 node moved 15869 keys (15.9%; ideal ~16.7%)
```
