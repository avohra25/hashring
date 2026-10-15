// Package hashring implements a consistent hashing ring with virtual nodes.
package hashring

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"sort"
	"strconv"
	"sync"
)

// ErrEmpty is returned when looking up a key in a ring with no nodes.
var ErrEmpty = errors.New("hashring: ring is empty")

// Ring maps keys to nodes using consistent hashing.
type Ring struct {
	mu       sync.RWMutex
	replicas int
	hashes   []uint32
	owners   map[uint32]string
	nodes    map[string]struct{}
}

// New creates a ring placing `replicas` virtual points per node.
func New(replicas int) *Ring {
	if replicas < 1 {
		replicas = 1
	}
	return &Ring{replicas: replicas, owners: map[uint32]string{}, nodes: map[string]struct{}{}}
}

// hash uses the first 4 bytes of MD5 for a well-spread (non-cryptographic) position.
func hash(s string) uint32 {
	sum := md5.Sum([]byte(s))
	return binary.BigEndian.Uint32(sum[:4])
}

// Add inserts nodes into the ring. Adding an existing node is a no-op.
func (r *Ring) Add(nodes ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, n := range nodes {
		if _, ok := r.nodes[n]; ok {
			continue
		}
		r.nodes[n] = struct{}{}
		for i := 0; i < r.replicas; i++ {
			h := hash(strconv.Itoa(i) + "#" + n)
			if _, taken := r.owners[h]; taken {
				continue // rare collision: keep first owner
			}
			r.owners[h] = n
			r.hashes = append(r.hashes, h)
		}
	}
	sort.Slice(r.hashes, func(i, j int) bool { return r.hashes[i] < r.hashes[j] })
}

// Remove deletes a node and its virtual points.
func (r *Ring) Remove(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.nodes[node]; !ok {
		return
	}
	delete(r.nodes, node)
	kept := r.hashes[:0]
	for _, h := range r.hashes {
		if r.owners[h] == node {
			delete(r.owners, h)
		} else {
			kept = append(kept, h)
		}
	}
	r.hashes = kept
}

func (r *Ring) search(key string) int {
	h := hash(key)
	i := sort.Search(len(r.hashes), func(i int) bool { return r.hashes[i] >= h })
	if i == len(r.hashes) {
		i = 0
	}
	return i
}

// Get returns the node owning key.
func (r *Ring) Get(key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.hashes) == 0 {
		return "", ErrEmpty
	}
	return r.owners[r.hashes[r.search(key)]], nil
}

// GetN returns up to n distinct nodes for key, walking clockwise (for replication).
func (r *Ring) GetN(key string, n int) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.hashes) == 0 {
		return nil, ErrEmpty
	}
	if n > len(r.nodes) {
		n = len(r.nodes)
	}
	out := make([]string, 0, n)
	seen := map[string]bool{}
	for i, start := 0, r.search(key); len(out) < n; i++ {
		o := r.owners[r.hashes[(start+i)%len(r.hashes)]]
		if !seen[o] {
			seen[o] = true
			out = append(out, o)
		}
	}
	return out, nil
}

// Nodes returns the sorted list of physical nodes.
func (r *Ring) Nodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.nodes))
	for n := range r.nodes {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
