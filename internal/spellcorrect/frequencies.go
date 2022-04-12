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

	"github.com/Saimunyz/speller/internal/mph"
)

type ngram []uint64

type Frequencies struct {
	MinWord int
	MinFreq int
	// UniGramProbs map[uint64]float64
	// Trie         *WordTrie
	Unigrams     []int32
	Bigrams      []int32
	Trigrams     []int32
	UnigramsHash *mph.Table
	BigramsHash  *mph.Table
	TrigramsHash *mph.Table
	TotalWords   int
}

// NewFrequencis - creates new Frequencies instance
func NewFrequencies(minWord, minFreq int) *Frequencies {
	ans := Frequencies{
		MinWord: minWord,
		MinFreq: minFreq,
		// UniGramProbs: make(map[uint64]float64),
		// Trie:         newWordTrie(0),
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
	o.Unigrams = data.Unigrams
	o.Bigrams = data.Bigrams
	o.Trigrams = data.Trigrams
	o.UnigramsHash = data.UnigramsHash
	o.BigramsHash = data.BigramsHash
	o.TrigramsHash = data.TrigramsHash
	o.TotalWords = data.TotalWords
	// o.Trie = data.Trie
	// o.UniGramProbs = data.UniGramProbs

	return nil
}

// TrainNgrams - traning ngrams model from big corpus
func (o *Frequencies) TrainNgrams(in io.Reader) error {
	// var hashes []uint64
	var (
		words      [][]string
		totalWords int
	)
	unigrams := make(map[string]int)
	bigrams := make(map[string]int)
	trigrams := make(map[string]int)

	// reads from file and counting freq of unigrams
	t := time.Now()
	scanner := bufio.NewScanner(in)
	// scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		var line []string
		rawLine := scanner.Text()
		splittedWords := strings.Fields(rawLine)
		for _, s := range splittedWords {
			s = strings.TrimRightFunc(s, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			word := strings.ToLower(s)

			totalWords++

			if len([]rune(word)) < o.MinWord {
				continue
			}
			line = append(line, word)
			unigrams[word]++
		}
		words = append(words, line)
	}

	// attempt to reduce memory allocation
	words = words[:len(words):len(words)]
	log.Println("time load tokens", time.Since(t), totalWords)
	t = time.Now()

	err := scanner.Err()
	if err != nil {
		return err
	}

	// free memory
	runtime.GC()

	o.TotalWords = totalWords

	for i := range words {
		bi := Ngrams(words[i], 2)
		for _, word := range bi {
			bigrams[word]++
		}
	}

	for i := range words {
		tri := Ngrams(words[i], 3)
		for _, word := range tri {
			trigrams[word]++
		}
	}

	unigramsSlice := make([]int32, len(unigrams))
	tmpUnigram := make([]string, 0, len(unigrams))
	for unigram := range unigrams {
		tmpUnigram = append(tmpUnigram, unigram)
	}
	hUnigrams := mph.Build(tmpUnigram)
	for unigram, freq := range unigrams {
		idx, _ := hUnigrams.Lookup(unigram)
		unigramsSlice[idx] = int32(freq)
	}

	bigramsSlice := make([]int32, len(bigrams))
	tmpBigram := make([]string, 0, len(bigrams))
	for bigram := range bigrams {
		tmpBigram = append(tmpBigram, bigram)
	}
	hBigrams := mph.Build(tmpBigram)
	for bigram, freq := range bigrams {
		idx, _ := hBigrams.Lookup(bigram)
		bigramsSlice[idx] = int32(freq)
	}

	trigramsSlice := make([]int32, len(trigrams))
	tmpTrigram := make([]string, 0, len(trigrams))
	for trigram := range trigrams {
		tmpTrigram = append(tmpTrigram, trigram)
	}
	hTrigrams := mph.Build(tmpTrigram)
	for trigram, freq := range trigrams {
		idx, _ := hTrigrams.Lookup(trigram)
		trigramsSlice[idx] = int32(freq)
	}

	// o.Trie = newWordTrie(totalWords)

	// counting unigrams probs
	// for k, v := range unigrams {
	// 	if v < o.MinFreq {
	// 		continue
	// 	} else {
	// 		o.UniGramProbs[k] = float64(v) / float64(totalWords)
	// 	}
	// }
	log.Println("time to load frequencies", time.Since(t))

	o.Unigrams = unigramsSlice
	o.Bigrams = bigramsSlice
	o.Trigrams = trigramsSlice

	o.UnigramsHash = hUnigrams
	o.BigramsHash = hBigrams
	o.TrigramsHash = hTrigrams

	return nil
}

func (o *Frequencies) GetUnigramFreq(word string) int32 {
	idx, ok := o.UnigramsHash.Lookup(word)
	if !ok {
		return 0
	}

	freq := o.Unigrams[idx]

	return freq
}

func (o *Frequencies) GetBigramFreq(word string) int32 {
	idx, ok := o.BigramsHash.Lookup(word)
	if !ok {
		return 0
	}

	freq := o.Bigrams[idx]

	return freq
}

func (o *Frequencies) GetTrigramFreq(word string) int32 {
	idx, ok := o.TrigramsHash.Lookup(word)
	if !ok {
		return 0
	}

	freq := o.Trigrams[idx]

	return freq
}

func (o *Frequencies) GetUnigramProb(word string) float64 {
	freq := o.GetUnigramFreq(word)
	if freq == 0 {
		return 0.0
	}

	prob := float64(freq) / float64(o.TotalWords)
	return prob
}

func (o *Frequencies) GetBigramProb(word string) float64 {
	bigramFreq := o.GetBigramFreq(word)
	if bigramFreq == 0 {
		return 0.0
	}

	idx := strings.Index(word, " ")
	unigramFreq := o.GetUnigramFreq(word[:idx])
	if unigramFreq == 0 {
		return 0.0
	}

	prob := float64(bigramFreq) / float64(unigramFreq)
	return prob
}

func (o *Frequencies) GetTrigramProb(word string) float64 {
	trigramFreq := o.GetTrigramFreq(word)
	if trigramFreq == 0 {
		return 0.0
	}

	idx := strings.LastIndex(word, " ")
	bigramFreq := o.GetBigramFreq(word[:idx])
	if bigramFreq == 0 {
		return 0.0
	}

	prob := float64(trigramFreq) / float64(bigramFreq)
	return prob
}

// Get - getter for frequencies of N-grams
// func (o *Frequencies) Get(tokens []string) float64 {

// 	// hashes := make([]uint64, len(tokens))
// 	// for i := range tokens {
// 	// 	hashes[i] = hashString(tokens[i])
// 	// }
// 	// if len(hashes) == 1 {
// 	// 	return o.UniGramProbs[hashes[0]]
// 	// }
// 	// node := o.Trie.search(hashes)
// 	// if node == nil {
// 	// 	return 0.0
// 	// }
// 	// return node.Prob
// }

// // Node - trie node
// type Node struct {
// 	Freq     int
// 	Prob     float64
// 	Children map[uint64]*Node
// }

// // newNode - creates new node instance
// func newNode(freq int) *Node {
// 	n := Node{
// 		Freq:     freq,
// 		Children: make(map[uint64]*Node),
// 	}
// 	return &n
// }

// type WordTrie struct {
// 	Root *Node
// }

// // newWordTrie - creates new WordTrie instance
// func newWordTrie(lenTokens int) *WordTrie {
// 	trie := WordTrie{
// 		Root: newNode(lenTokens),
// 	}
// 	return &trie
// }

// // put - puts ngram in trie
// func (o *WordTrie) put(key ngram) {
// 	//The assumption that we first add the 1gram then the 2gram etc is made
// 	current := o.Root
// 	var i int
// 	for ; i < len(key)-1; i++ {
// 		current = current.Children[key[i]]
// 	}
// 	node, ok := current.Children[key[i]]
// 	if ok {
// 		node.Freq++
// 	} else {
// 		node = newNode(1)
// 		current.Children[key[i]] = node
// 	}
// 	node.Prob = float64(node.Freq) / float64(current.Freq)
// }

// // search - looking for ngrams in trie
// func (o *WordTrie) search(key ngram) *Node {
// 	tmp := o.Root
// 	for i := 0; i < len(key); i++ {
// 		if next, ok := tmp.Children[key[i]]; ok {
// 			tmp = next
// 		} else {
// 			return nil
// 		}
// 	}
// 	return tmp
// }

// // hashString - hashes string
// func hashString(s string) uint64 {
// 	return fnv1a.HashString64(s)
// }

// Ngrams - returns ngrams from tokens
func Ngrams(words []string, size int) []string {
	outCap := len(words) - (size - 1)
	if outCap < 0 {
		outCap = 0
	}
	out := make([]string, 0, outCap)
	for i := 0; i+size <= len(words); i++ {
		out = append(out, strings.Join(words[i:i+size:i+size], " "))
	}
	return out
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

// // ngrams - returns channel with ngrams
// func ngrams(words []uint64, size int) <-chan ngram {
// 	out := make(chan ngram)
// 	go func() {
// 		defer close(out)
// 		for i := 0; i+size <= len(words); i++ {
// 			out <- words[i : i+size : i+size]
// 		}
// 	}()
// 	return out
// }
