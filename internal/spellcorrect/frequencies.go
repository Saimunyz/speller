package spellcorrect

import (
	"bufio"
	"compress/gzip"
	"encoding/gob"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/segmentio/fasthash/fnv1a"
)

type ngram []uint64

type Frequencies struct {
	MinWord      int
	MinFreq      int
	UniGramProbs map[uint64]float64
	Trie         *WordTrie
}

// NewFrequencis - creates new Frequencies instance
func NewFrequencies(minWord, minFreq int) *Frequencies {
	ans := Frequencies{
		MinWord:      minWord,
		MinFreq:      minFreq,
		UniGramProbs: make(map[uint64]float64),
		Trie:         newWordTrie(0),
	}
	return &ans
}

// SaveModel - saves trained speller model
func (o *Frequencies) SaveModel(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := gzip.NewWriter(f)
	defer w.Close()

	runtime.GC()

	enc := gob.NewEncoder(w)
	err = enc.Encode(o)
	if err != nil {
		return err
	}

	runtime.GC()

	return nil
}

// LoadModel - loades trained speller model from file
func (o *Frequencies) LoadModel(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	var data Frequencies

	dec := gob.NewDecoder(gz)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}

	o.MinFreq = data.MinFreq
	o.MinWord = data.MinWord
	o.Trie = data.Trie
	o.UniGramProbs = data.UniGramProbs

	return nil
}

// TrainNgramsOnline - attempt to make real-time traning
func (o *Frequencies) TrainNgramsOnline(tokens []string) error {
	var hashes []uint64

	o.Trie.Root.Freq += len(tokens)

	// for _, query := range queries

	for _, token := range tokens {
		hashes = append(hashes, hashString(token))
	}

	for i := 1; i < 4; i++ {
		grams := ngrams(hashes, i)
		for _ngram := range grams {
			o.Trie.put(_ngram)
		}
	}

	// update unigrams probs
	for i := range hashes {
		// need to update root.Freq and whole trie after this?
		node := o.Trie.search([]uint64{hashes[i]})
		o.UniGramProbs[hashes[i]] = float64(node.Freq) / float64(o.Trie.Root.Freq)
	}

	for hash := range o.UniGramProbs {
		node := o.Trie.search([]uint64{hash})
		o.UniGramProbs[hash] = float64(node.Freq) / float64(o.Trie.Root.Freq)
		node.Prob = float64(node.Freq) / float64(o.Trie.Root.Freq)

		if node.Children != nil {
			go o.updateChield(node)
		}
	}

	return nil
}

func (o *Frequencies) updateChield(parent *Node) {
	for _, child := range parent.Children {
		if child != nil {
			child.Prob = float64(child.Freq) / float64(parent.Freq)
			if child.Children != nil {
				go o.updateChield(child)
			}
		}
	}
}

// TrainNgrams - traning ngrams model from big corpus
func (o *Frequencies) TrainNgrams(in io.Reader) error {
	if len(o.UniGramProbs) != 0 {
		return nil
	}

	// var hashes []uint64
	var (
		hashes     [][]uint64
		totalWords int
	)
	unigrams := make(map[uint64]int)
	bl := make(map[uint64]bool)

	// reads from file and counting freq of unigrams
	t := time.Now()
	scanner := bufio.NewScanner(in)
	// scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		var lineHashes []uint64
		rawLine := scanner.Text()
		splittedWords := strings.Fields(rawLine)
		for _, s := range splittedWords {
			s = strings.TrimRightFunc(s, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			word := strings.ToLower(s)

			totalWords++

			if len([]rune(word)) < o.MinWord {
				// bl[lineHashes[len(lineHashes)-1]] = true
				continue
			}
			lineHashes = append(lineHashes, hashString(word))
			unigrams[lineHashes[len(lineHashes)-1]]++
		}
		hashes = append(hashes, lineHashes)
	}

	// attempt to reduce memory allocation
	hashes = hashes[:len(hashes):len(hashes)]
	log.Println("time load tokens", time.Since(t), totalWords)
	t = time.Now()

	err := scanner.Err()
	if err != nil {
		return err
	}

	// free memory
	runtime.GC()

	o.Trie = newWordTrie(totalWords)

	// counting unigrams probs
	for k, v := range unigrams {
		if v < o.MinFreq {
			bl[k] = true
		} else {
			o.UniGramProbs[k] = float64(v) / float64(totalWords)
		}
	}

	// counting N-grams probs and store them in trie
	for i := 1; i < 4; i++ {
		for _, h := range hashes {
			grams := ngrams(h, i)
			for _ngram := range grams {
				add := true
				for j := range _ngram {
					if bl[_ngram[j]] {
						add = false
						break
					}
				}
				if add {
					o.Trie.put(_ngram)
				}
			}
		}
	}
	log.Println("time to load frequencies", time.Since(t))

	return nil
}

// Get - getter for frequencies of N-grams
func (o *Frequencies) Get(tokens []string) float64 {
	hashes := make([]uint64, len(tokens))
	for i := range tokens {
		hashes[i] = hashString(tokens[i])
	}
	if len(hashes) == 1 {
		return o.UniGramProbs[hashes[0]]
	}
	node := o.Trie.search(hashes)
	if node == nil {
		return 0.0
	}
	return node.Prob
}

// Node - trie node
type Node struct {
	Freq     int
	Prob     float64
	Children map[uint64]*Node
}

// newNode - creates new node instance
func newNode(freq int) *Node {
	n := Node{
		Freq:     freq,
		Children: make(map[uint64]*Node),
	}
	return &n
}

type WordTrie struct {
	Root *Node
}

// newWordTrie - creates new WordTrie instance
func newWordTrie(lenTokens int) *WordTrie {
	trie := WordTrie{
		Root: newNode(lenTokens),
	}
	return &trie
}

// put - puts ngram in trie
func (o *WordTrie) put(key ngram) {
	//The assumption that we first add the 1gram then the 2gram etc is made
	current := o.Root
	var i int
	for ; i < len(key)-1; i++ {
		current = current.Children[key[i]]
	}
	node, ok := current.Children[key[i]]
	if ok {
		node.Freq++
	} else {
		node = newNode(1)
		current.Children[key[i]] = node
	}
	node.Prob = float64(node.Freq) / float64(current.Freq)
}

// search - looking for ngrams in trie
func (o *WordTrie) search(key ngram) *Node {
	tmp := o.Root
	for i := 0; i < len(key); i++ {
		if next, ok := tmp.Children[key[i]]; ok {
			tmp = next
		} else {
			return nil
		}
	}
	return tmp
}

// hashString - hashes string
func hashString(s string) uint64 {
	return fnv1a.HashString64(s)
}

// TokenNgrams - returns ngrams from tokens
func TokenNgrams(words []string, size int) [][]string {
	outCap := len(words) - (size - 1)
	if outCap < 0 {
		outCap = 0
	}
	out := make([][]string, 0, outCap)
	for i := 0; i+size <= len(words); i++ {
		out = append(out, words[i:i+size:i+size])
	}
	return out
}

// ngrams - returns channel with ngrams
func ngrams(words []uint64, size int) <-chan ngram {
	out := make(chan ngram)
	go func() {
		defer close(out)
		for i := 0; i+size <= len(words); i++ {
			out <- words[i : i+size : i+size]
		}
	}()
	return out
}
