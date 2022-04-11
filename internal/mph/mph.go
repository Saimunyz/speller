// Package mph implements a minimal perfect hash table over strings.
package mph

import "sort"

// A Table is an immutable hash table that provides constant-time lookups of key
// indices using a minimal perfect hash.
type Table struct {
	Keys       []string
	Level0     []uint32 // power of 2 size
	Level0Mask int      // len(Level0) - 1
	Level1     []uint32 // power of 2 size >= len(keys)
	Level1Mask int      // len(Level1) - 1
}

// Build builds a Table from keys using the "Hash, displace, and compress"
// algorithm described in http://cmph.sourceforge.net/papers/esa09.pdf.
func Build(keys []string) *Table {
	var (
		level0        = make([]uint32, nextPow2(len(keys)/4))
		level0Mask    = len(level0) - 1
		level1        = make([]uint32, nextPow2(len(keys)))
		level1Mask    = len(level1) - 1
		sparseBuckets = make([][]int, len(level0))
		zeroSeed      = murmurSeed(0)
	)
	for i, s := range keys {
		n := int(zeroSeed.hash(s)) & level0Mask
		sparseBuckets[n] = append(sparseBuckets[n], i)
	}
	var buckets []indexBucket
	for n, vals := range sparseBuckets {
		if len(vals) > 0 {
			buckets = append(buckets, indexBucket{n, vals})
		}
	}
	sort.Sort(bySize(buckets))

	occ := make([]bool, len(level1))
	var tmpOcc []int
	for _, bucket := range buckets {
		var seed murmurSeed
	trySeed:
		tmpOcc = tmpOcc[:0]
		for _, i := range bucket.vals {
			n := int(seed.hash(keys[i])) & level1Mask
			if occ[n] {
				for _, n := range tmpOcc {
					occ[n] = false
				}
				seed++
				goto trySeed
			}
			occ[n] = true
			tmpOcc = append(tmpOcc, n)
			level1[n] = uint32(i)
		}
		level0[int(bucket.n)] = uint32(seed)
	}

	return &Table{
		Keys:       keys,
		Level0:     level0,
		Level0Mask: level0Mask,
		Level1:     level1,
		Level1Mask: level1Mask,
	}
}

func nextPow2(n int) int {
	for i := 1; ; i *= 2 {
		if i >= n {
			return i
		}
	}
}

// Lookup searches for s in t and returns its index and whether it was found.
func (t *Table) Lookup(s string) (n uint32, ok bool) {
	i0 := int(murmurSeed(0).hash(s)) & t.Level0Mask
	seed := t.Level0[i0]
	i1 := int(murmurSeed(seed).hash(s)) & t.Level1Mask
	n = t.Level1[i1]
	return n, s == t.Keys[int(n)]
}

type indexBucket struct {
	n    int
	vals []int
}

type bySize []indexBucket

func (s bySize) Len() int           { return len(s) }
func (s bySize) Less(i, j int) bool { return len(s[i].vals) > len(s[j].vals) }
func (s bySize) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
