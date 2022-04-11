package spellcorrect

import (
	"strings"
	"testing"
)

func TestFrequencies(t *testing.T) {
	tokens := []string{"I", "program", "go", "I", "code", "and", "I", "cook", "code"}
	in := strings.NewReader(strings.Join(tokens, " "))
	freq := NewFrequencies(0, 0)
	if err := freq.TrainNgrams(in); err != nil {
		t.Errorf(err.Error())
		return
	}

	if prob := freq.GetUnigramProb("i"); prob < 0.3 || prob > 0.34 {
		t.Errorf("unigram prob wrong")
		return
	}

	if prob := freq.GetBigramProb("i code"); prob < 0.3 || prob > 0.34 {
		t.Errorf("bigram prob wrong")
		return
	}

	if prob := freq.GetTrigramProb("i program go"); prob < 0.99 || prob > 1 {
		t.Errorf("trigram prob wrong")
		return
	}

}

// func TestWordTrie(t *testing.T) {
// 	words := []uint64{
// 		1, 2, 3, 4, 5, 6, 1, 2,
// 	}

// 	trie := newWordTrie(len(words))

// 	unigrams := ngrams(words, 1)
// 	for unigram := range unigrams {
// 		trie.put(unigram)
// 	}

// 	s := ngram{uint64(2)}
// 	if n := trie.search(s); n.Freq != 2 {
// 		t.Errorf("error computing freq")
// 		return
// 	}

// 	if n := trie.search(ngram{uint64(79)}); n != nil {
// 		t.Errorf("error searching not existant")
// 		return
// 	}
// 	bigrams := ngrams(words, 2)
// 	for bigram := range bigrams {
// 		trie.put(bigram)
// 	}

// 	if n := trie.search(ngram{uint64(1)}); n.Freq != 2 {
// 		t.Errorf("error computing freq")
// 		return
// 	}
// 	if n := trie.search(ngram{uint64(1), uint64(2)}); n.Freq != 2 {
// 		t.Errorf("error computing freq")
// 		return
// 	}
// }
