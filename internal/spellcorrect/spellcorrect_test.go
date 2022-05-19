package spellcorrect

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/Saimunyz/speller/internal/config"
	"github.com/Saimunyz/speller/internal/spell"
)

func getSpellCorrector() *SpellCorrector {
	tokenizer := NewSimpleTokenizer()
	cfg := &config.Config{
		SpellerConfig: config.SpellerConfig{MinWordFreq: 1, MinWordLength: 5, AutoTrainMode: false, Penalty: 1},
	}
	freq := NewFrequencies(cfg)
	sc := NewSpellCorrector(tokenizer, freq, []float64{100, 15, 5}, cfg)
	return sc
}

func TestTrain(t *testing.T) {
	trainwords := "golang 100\ngoland 1"
	traindata := `golang python C erlang golang java java golang goland`
	r := strings.NewReader(traindata)
	r1 := strings.NewReader(trainwords)

	sc := getSpellCorrector()
	dict := FreqDicts{Name: "default", Reader: r1}
	if err := sc.Train(r, dict); err != nil {
		t.Errorf(err.Error())
		return
	}
	if prob := sc.frequencies.Get([]string{"golang"}); prob > 0.34 && prob < 0.33 {
		t.Errorf("invalid prob %f", prob)
		return
	}

	suggestions, _ := sc.spell.Lookup("gola", spell.SuggestionLevel(spell.LevelAll))
	if len(suggestions) != 2 {
		t.Errorf("calculated wrong suggestions")
		return
	}
	expected := map[string]bool{
		"golang": true, "goland": true,
	}
	for i := range suggestions {
		if !expected[suggestions[i].Word] {
			t.Errorf("missing suggestion")
			return
		}
	}
}

func BenchmarkProduct(b *testing.B) {
	left := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	right := [][]string{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}, {"6"}, {"7"}, {"8"}, {"9"}, {"10"}}
	l := NewWordWithDistList(left)
	r := make([]WordsWithDistList, len(right))
	for i := range r {
		r[i] = NewWordWithDistList(right[i])
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		product(l, r)
	}
}

func BenchmarkCombos(b *testing.B) {
	input := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

	tokens := NewWordWithDistList(input)

	var in []WordsWithDistList
	for i := range tokens {
		var sug WordsWithDistList
		if i != 0 && i < len(tokens)/2 {
			for k := 1; k <= i; k++ {
				sug = append(sug, NewWordWithDist(strings.Repeat(fmt.Sprintf("%d", i), k), 0))
			}
		}
		if len(sug) == 0 {
			sug = append(sug, tokens[i])
		}
		in = append(in, sug)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = combos(in)
	}
}

func TestSpellCorrect(t *testing.T) {
	trainwords := "golang 100\ngoland 1"
	traindata := `golang python C erlang golang java java golang goland`
	r := strings.NewReader(traindata)
	r1 := strings.NewReader(trainwords)

	sc := getSpellCorrector()
	dict := FreqDicts{Name: "default", Reader: r1}
	if err := sc.Train(r, dict); err != nil {
		t.Errorf(err.Error())
		return
	}

	s1 := "restaurant in Bonn"
	tokens, _ := sc.Tokenizer.Tokens(strings.NewReader(s1))
	suggestions := sc.SpellCorrect(tokens)
	for i, sug := range suggestions {
		if sug.Tokens == nil {
			suggestions = suggestions[:i]
			break
		}
	}
	if len(suggestions) != 1 {
		t.Errorf("error getting suggestion for not existant")
		return
	}
	if suggestions[0].score != math.Inf(-1) {
		t.Errorf("error getting suggestion for not existant (different)")
		return
	}
}

func TestGetSuggestionCandidates(t *testing.T) {
	input := []string{"1", "2", "3"}

	tokens := NewWordWithDistList(input)

	sugMap := map[int]spell.SuggestionList{
		0: {
			spell.Suggestion{Distance: 1, Entry: spell.Entry{Word: "a"}},
			spell.Suggestion{Distance: 2, Entry: spell.Entry{Word: "aa"}},
		},
		1: {
			spell.Suggestion{Distance: 1, Entry: spell.Entry{Word: "b"}},
		},
		2: {},
	}

	var allSuggestions []WordsWithDistList
	for i := range tokens {
		allSuggestions = append(allSuggestions, nil)
		allSuggestions[i] = append(allSuggestions[i], tokens[i])
		suggestions := sugMap[i]
		for j := 0; j < len(suggestions) && j < 10; j++ {
			allSuggestions[i] = append(allSuggestions[i], NewWordWithDist(suggestions[j].Word, 0))
		}

	}

	tmp := [][]string{

		{"aa", "b", "3"},
		{"aa", "2", "3"},
		{"a", "b", "3"},
		{"a", "2", "3"},
		{"1", "b", "3"},
		{"1", "2", "3"},
	}

	expected := make([]WordsWithDistList, len(tmp))
	for i := range expected {
		expected[i] = NewWordWithDistList(tmp[i])
	}

	sc := getSpellCorrector()

	candidates := sc.getSuggestionCandidates(allSuggestions)

	for i, sug := range candidates {
		if sug.Tokens == nil {
			candidates = candidates[:i]
			break
		}
	}

	if len(candidates) != len(expected) {
		t.Errorf("invalid length")
		return
	}

	expect := make(map[uint64]bool)
	for i := range expected {
		expect[hashTokens(expected[i])] = true
	}
	for i := range candidates {
		if !expect[hashTokens(NewWordWithDistList(candidates[i].Tokens))] {
			t.Errorf("%v not in expected", candidates[i].Tokens)
			return
		}
	}
}
