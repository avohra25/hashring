package hashring

import (
	"strconv"
	"testing"
)

func TestEmpty(t *testing.T) {
	if _, err := New(10).Get("x"); err != ErrEmpty {
		t.Fatalf("want ErrEmpty, got %v", err)
	}
}

func TestDeterministic(t *testing.T) {
	a, b := New(50), New(50)
	a.Add("a", "b", "c")
	b.Add("c", "a", "b")
	for i := 0; i < 1000; i++ {
		k := strconv.Itoa(i)
		x, _ := a.Get(k)
		y, _ := b.Get(k)
		if x != y {
			t.Fatalf("key %s: %s != %s", k, x, y)
		}
	}
}

func TestMinimalRemapOnAdd(t *testing.T) {
	r := New(160)
	r.Add("n0", "n1", "n2", "n3")
	const N = 20000
	before := make([]string, N)
	for i := range before {
		before[i], _ = r.Get(strconv.Itoa(i))
	}
	r.Add("n4")
	moved := 0
	for i, old := range before {
		now, _ := r.Get(strconv.Itoa(i))
		if now != old {
			if now != "n4" {
				t.Fatalf("key moved to existing node %s", now)
			}
			moved++
		}
	}
	if frac := float64(moved) / N; frac < 0.1 || frac > 0.3 {
		t.Fatalf("moved fraction %.3f outside [0.1,0.3]", frac)
	}
}

func TestRemove(t *testing.T) {
	r := New(100)
	r.Add("a", "b", "c")
	r.Remove("b")
	for i := 0; i < 1000; i++ {
		if n, _ := r.Get(strconv.Itoa(i)); n == "b" {
			t.Fatal("removed node still owns keys")
		}
	}
	if got := r.Nodes(); len(got) != 2 {
		t.Fatalf("nodes = %v", got)
	}
}

func TestBalance(t *testing.T) {
	r := New(200)
	for i := 0; i < 8; i++ {
		r.Add("node" + strconv.Itoa(i))
	}
	c := map[string]int{}
	const N = 80000
	for i := 0; i < N; i++ {
		n, _ := r.Get("k" + strconv.Itoa(i))
		c[n]++
	}
	for n, v := range c {
		if v < N/8*7/10 || v > N/8*13/10 {
			t.Errorf("%s has %d keys, too far from %d", n, v, N/8)
		}
	}
}

func TestGetN(t *testing.T) {
	r := New(50)
	r.Add("a", "b", "c")
	got, _ := r.GetN("key", 5)
	if len(got) != 3 {
		t.Fatalf("want 3 distinct nodes, got %v", got)
	}
	first, _ := r.Get("key")
	if got[0] != first {
		t.Fatalf("GetN[0]=%s, Get=%s", got[0], first)
	}
}

func BenchmarkGet(b *testing.B) {
	r := New(160)
	for i := 0; i < 16; i++ {
		r.Add("n" + strconv.Itoa(i))
	}
	for i := 0; i < b.N; i++ {
		r.Get(strconv.Itoa(i))
	}
}
