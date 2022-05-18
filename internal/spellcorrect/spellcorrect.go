package spellcorrect

import (
	"bufio"
	"io"
	"log"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Saimunyz/speller/internal/spell"
	"github.com/segmentio/fasthash/fnv1a"
)

type Suggestion struct {
	score  float64
	Tokens []string
}

// FrequencyContainer - all the necessary functions for working with the frequency layer
type FrequencyContainer interface {
	TrainNgrams(in io.Reader) error
	Get(tokens []string) float64
	LoadModel(filename string) error
	SaveModel(filename string) error
}

// Tokinizer - tokenizer function from token layer
type Tokenizer interface {
	Tokens(in io.Reader) ([]string, error)
}

// freqDicts - for holding dictionary names and readers
type FreqDicts struct {
	Name   string
	Reader io.Reader
}

type SpellCorrector struct {
	Tokenizer     Tokenizer
	frequencies   FrequencyContainer
	spell         *spell.Spell
	weights       []float64
	minFreq       int
	penalty       float64
	autoTrainMode bool
}

// NewSpellCorrector - creates new SpellCorrector instance
func NewSpellCorrector(
	tokenizer Tokenizer,
	frequencies FrequencyContainer,
	weights []float64,
	autoTrainMode bool,
	minFreq int,
	penalty float64,
) *SpellCorrector {
	ans := SpellCorrector{
		Tokenizer:     tokenizer,
		frequencies:   frequencies,
		spell:         spell.New(),
		weights:       weights,
		minFreq:       minFreq,
		penalty:       penalty,
		autoTrainMode: autoTrainMode,
	}
	ans.spell.MaxEditDistance = 3
	ans.spell.PrefixLength = 8
	return &ans
}

// SaveModel - saves trained speller model
func (o *SpellCorrector) SaveModel(filename string) error {
	err := o.frequencies.SaveModel(filename)
	if err != nil {
		return err
	}
	return nil
}

// LoadModel - loades trained speller model from file
func (o *SpellCorrector) LoadModel(filename string) error {
	err := o.frequencies.LoadModel(filename)
	if err != nil {
		return err
	}
	return nil
}

// loadFreqDict - loads ferequencies dictionary in spell lib
func (o *SpellCorrector) LoadFreqDict(in io.Reader, dictName string) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		freq, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		}

		if freq < uint64(o.minFreq) {
			continue
		}

		o.spell.AddEntry(spell.Entry{
			Frequency: freq,
			Word:      parts[0],
		},
			spell.DictionaryName(dictName),
		)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	o.spell.PrefixLength = 7 // return to default value

	return nil
}

// Train - train n-grams model from in Reader and read freq from in2 Reader
func (o *SpellCorrector) Train(sentences io.Reader, dicts ...FreqDicts) error {
	// counting n-grams freq
	err := o.frequencies.TrainNgrams(sentences)
	if err != nil {
		return err
	}
	runtime.GC()

	// ToDo refactoring for better reading
	// load freq dict
	// assume than first would be common dictionary, second shortWords
	// all names declared in o.freqDicts
	t := time.Now()
	for _, dict := range dicts {
		err = o.LoadFreqDict(dict.Reader, dict.Name)
		if err != nil {
			return err
		}
	}

	log.Printf("Freq dictionary loaded: %v", time.Since(t))

	runtime.GC()
	return nil
}

// hashTokens - returns hash of given token
func hashTokens(tokens []string) uint64 {
	h := fnv1a.Init64
	for i := range tokens {
		h = fnv1a.AddString64(h, tokens[i])
	}
	return h
}

// product - computes product() of given slice
func product(a []string, b [][]string) [][]string {
	size := len(a) * len(b)
	items := make([][]string, size)

	var k int
	for i := range a {
		for j := range b {
			var h int
			items[k] = make([]string, len(b[0])+1)
			for _, word := range a[i : i+1 : i+1] {
				items[k][h] = word
				h++
			}
			for _, word := range b[j] {
				items[k][h] = word
				h++
			}
			k++
		}
	}
	return items
}

// sliceToSliceOfSlice - change slice to slice of slice
func sliceToSliceOfSlice(words []string) [][]string {
	res := make([][]string, len(words))
	for i := range res {
		res[i] = words[i : i+1]
	}

	return res
}

// combos - permutation of all sentences
func combos(in [][]string) [][]string {
	tmpP := sliceToSliceOfSlice(in[len(in)-1])

	for i := len(in) - 2; i >= 0; i-- {
		tmpP = product(in[i], tmpP)
	}
	return tmpP
}

func (o *SpellCorrector) CheckInFreqDict(query string) bool {
	return o.spell.CheckExistance(query)
}

// lookupTokens - finds all the suggestions given by the spell library and takes the top 20 of them
func (o *SpellCorrector) lookupTokens(tokens []string) ([][]string, map[string]float64) {
	const amountOfSuggestions = 10
	allSuggestions := make([][]string, len(tokens))
	dist := make(map[string]float64, len(tokens))

	for i := range tokens {
		// dont look at short words
		if len([]rune(tokens[i])) < 2 || strings.ContainsAny(tokens[i], "1234567890") {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
			dist[tokens[i]] = 0
			continue
		}

		// gets suggestions
		var suggestions spell.SuggestionList
		o.spell.MaxEditDistance = 2

		suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		if len(suggestions) == 0 {
			o.spell.MaxEditDistance = 3
			suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelClosest))
			if len(suggestions) == 0 {
				suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
			}
		}
		// gets 5 first suggestions
		if len(allSuggestions[i]) == 0 {
			// size := amountOfSuggestions
			// if size > len(suggestions) {
			// 	size = len(suggestions)
			// }
			// allSuggestions[i] = make([]string, size)

			// var wordExist bool
			// if len(suggestions) > 0 && suggestions[0].Distance == 0 {
			// 	wordExist = true
			// }

			for j := 0; j < len(suggestions) && j < amountOfSuggestions; j++ {
				// if wordExist && suggestions[j].Distance > 1 {
				// 	continue
				// }

				allSuggestions[i] = append(allSuggestions[i], suggestions[j].Word)
				if _, ok := dist[suggestions[j].Word]; !ok {
					dist[suggestions[j].Word] = float64(suggestions[j].Distance) + float64(j)/2*o.penalty
				}
			}
		}
		// if no suggestions returns token
		if len(allSuggestions[i]) == 0 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
			dist[tokens[i]] = 0
		}
	}

	return allSuggestions, dist
}

// lookupTokens - finds all the suggestions given by the spell library and takes the top 20 of them
func (o *SpellCorrector) lookupTokens2(tokens []string) ([][]string, map[string]float64) {
	const amountOfSuggestions = 5
	allSuggestions := make([][]string, len(tokens))
	dist := make(map[string]float64, len(tokens))

	for i := range tokens {
		// dont look at short words
		if len([]rune(tokens[i])) < 3 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
			dist[tokens[i]] = 0
		}

		// gets suggestions
		var suggestions spell.SuggestionList
		o.spell.MaxEditDistance = 1

		suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelClosest))
		if len(suggestions) < 4 {
			o.spell.MaxEditDistance = 2
			suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		}
		// gets 5 first suggestions
		if len(allSuggestions[i]) == 0 {
			size := amountOfSuggestions
			if size > len(suggestions) {
				size = len(suggestions)
			}
			allSuggestions[i] = make([]string, size)

			for j := 0; j < len(suggestions) && j < amountOfSuggestions; j++ {
				allSuggestions[i][j] = suggestions[j].Word
				if _, ok := dist[suggestions[j].Word]; !ok {
					dist[suggestions[j].Word] = float64(suggestions[j].Distance) + float64(j)*o.penalty
				}
			}
		}
		// if no suggestions returns token
		if len(allSuggestions[i]) == 0 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
			dist[tokens[i]] = 0
		}
	}

	return allSuggestions, dist
}

// getInsertPosition - returns the position sorted in descending order
func getInsertPosition(nums []Suggestion, target Suggestion) int {
	min := 0
	max := len(nums) - 1
	for min <= max {
		mid := min + (max-min)/2

		switch {
		case target.score >= nums[mid].score:
			max = mid - 1
		case target.score < nums[mid].score:
			min = mid + 1
		}
	}
	return min
}

// insertPosition - insert value at pos, moving all others down
func insertPosition(suggesses []Suggestion, pos int, sugges Suggestion) {
	if pos >= len(suggesses) {
		return
	}
	if suggesses[pos].Tokens == nil {
		suggesses[pos] = sugges
		return
	}

	after := suggesses[pos:]

	// insert from the end
	for i, j := len(suggesses)-1, len(suggesses)-len(suggesses[:pos])-2; i > pos && j >= 0; i-- {
		suggesses[i] = after[j]
		j--
	}

	suggesses[pos] = sugges
}

func newSuggestions() []Suggestion {
	suggestions := make([]Suggestion, 10)

	for i := range suggestions {
		suggestions[i].score = math.Inf(-1)
	}

	return suggestions
}

// getSuggestionCandidates - returns slice of fixed typos with context N-grams
func (o *SpellCorrector) getSuggestionCandidates(allSuggestions [][]string, dist map[string]float64) []Suggestion {
	// combine suggestions
	suggestionStrings := combos(allSuggestions)
	seen := make(map[uint64]struct{}, len(suggestionStrings))
	suggestions := newSuggestions()
	for i := range suggestionStrings {
		// sugTokens := strings.Split(suggestionStrings[i], " ")
		h := hashTokens(suggestionStrings[i])
		if _, ok := seen[h]; !ok {
			seen[h] = struct{}{}
			sugges := Suggestion{
				score:  o.score(suggestionStrings[i], dist),
				Tokens: suggestionStrings[i],
			}
			pos := getInsertPosition(suggestions, sugges)
			insertPosition(suggestions, pos, sugges)
		}
	}

	return suggestions
}

// SpellCorrect - returns suggestions
func (o *SpellCorrector) SpellCorrect(tokens []string) []Suggestion {
	// tokens, _ := o.Tokenizer.Tokens(strings.NewReader(s))
	allSuggestions, dist := o.lookupTokens(tokens)
	items := o.getSuggestionCandidates(allSuggestions, dist)

	return items
}

// SpellCorrect - returns suggestions
func (o *SpellCorrector) SpellCorrect2(s string) []Suggestion {
	tokens, _ := o.Tokenizer.Tokens(strings.NewReader(s))
	allSuggestions, dist := o.lookupTokens2(tokens)
	items := o.getSuggestionCandidates(allSuggestions, dist)

	return items
}

func (o *SpellCorrector) SpellCorrectWithoutContext(s string) []string {
	if len([]rune(s)) < 2 {
		return []string{s}
	}

	// ToDo need to refactor?
	const shortWordsDict = "shortWords"

	suggestions, _ := o.spell.Lookup(s, spell.SuggestionLevel(spell.LevelClosest), spell.DictionaryOpts(
		spell.DictionaryName(shortWordsDict)),
	)
	result := make([]string, len(suggestions))

	for i := range result {
		result[i] = suggestions[i].Word
	}

	if len(result) < 1 {
		return []string{s}
	}

	return result
}

// getPenalty - returns penalty as a percentage of the
// obtained probability for the distance of the word
func getPenalty(prob float64, dist float64) float64 {
	var alpha float64

	if dist == 0 {
		return 0
	}

	// change space from 0 - 5 to 1 - 10
	relative := (dist - 0) / (5 - 0)
	// scaled_value := 1 + (10-1)*relative

	// alpha = math.Log10(scaled_value) * 100
	alpha = 0 + (100-0)*relative

	// if alpha >= 100. {
	// 	alpha = 99.
	// }

	prob = prob * float64(alpha) / float64(100)
	if prob < 0 {
		prob *= -1
	}
	return prob
}

// GetUnigram - returns unigram with penalties
func (o *SpellCorrector) GetUnigram(tokens []string) float64 {
	unigrams := tokens[:1:1]

	prob := o.frequencies.Get(unigrams)

	return prob
}

// GetBigram - returns bigrams
func (o *SpellCorrector) GetBigram(tokens []string) float64 {
	bigrams := tokens[:2:2]

	prob := o.frequencies.Get(bigrams)

	return prob
}

// GetTrigram - returns trigrams
func (o *SpellCorrector) GetTrigram(tokens []string) float64 {
	trigrams := tokens[:3:3]

	prob := o.frequencies.Get(trigrams)

	return prob
}

// calculateBigramScore - returns bigram score of a given words
func (o *SpellCorrector) calculateBigramScore(ngrams []string, dist map[string]float64) float64 {
	var score float64

	// penalty := len(bigrams)

	for i := 0; i+2 <= len(ngrams); i++ {
		bigrams := ngrams[i : i+2 : i+2]

		bigram := o.frequencies.Get(bigrams)
		if bigram != 0 {
			bigram -= getPenalty(bigram, dist[bigrams[0]]+dist[bigrams[1]])

			unigram := o.GetUnigram(bigrams)
			if unigram != 0 {
				unigram += o.weights[0]
				unigram -= getPenalty(unigram, dist[bigrams[0]])
			}

			score += unigram + bigram
		} else {
			tmp := o.calculateUnigramScore(bigrams, dist)
			score += (tmp + tmp)
		}
	}
	return score
}

// calculateUnigramScore - returns unigram score of a given words
func (o *SpellCorrector) calculateUnigramScore(ngrams []string, dist map[string]float64) float64 {
	var score float64

	penalty := len(ngrams)

	for i := 0; i+1 <= len(ngrams); i++ {
		unigrams := ngrams[i : i+1 : i+1]

		unigram := o.frequencies.Get(unigrams)
		if unigram != 0 {
			penalty--
			unigram -= getPenalty(unigram, dist[unigrams[0]])
		}

		score += unigram
	}

	if penalty > 0 {
		score = o.applyPenalty(score, penalty)
	}
	return score
}

// calculateTrigramScore -  returns trigrams score of a given words
func (o *SpellCorrector) calculateTrigramScore(ngrams []string, dist map[string]float64) float64 {
	var score float64

	for i := 0; i+3 <= len(ngrams); i++ {
		trigrams := ngrams[i : i+3 : i+3]

		trigram := o.frequencies.Get(trigrams)
		if trigram != 0 {
			trigram -= getPenalty(trigram, dist[trigrams[0]]+dist[trigrams[1]]+dist[trigrams[2]])

			bigram := o.GetBigram(trigrams)
			if bigram != 0 {
				bigram += o.weights[1]

				bigram -= getPenalty(bigram, dist[trigrams[0]]+dist[trigrams[1]])
			}
			unigram := o.GetUnigram(trigrams)
			if unigram != 0 {
				unigram += o.weights[0]
				unigram -= getPenalty(unigram, dist[trigrams[0]])
			}

			score += unigram + bigram + trigram
		} else {
			tmp := o.calculateBigramScore(trigrams, dist)
			score += tmp + tmp
		}
	}
	return score
}

// applyPenalty - applies penalty to given score
func (o *SpellCorrector) applyPenalty(score float64, penalty int) float64 {
	newScore := score
	for i := 0; i < penalty; i++ {
		newScore += score
	}
	return newScore
}

// score - scoring each sentence
func (o *SpellCorrector) score(tokens []string, dist map[string]float64) float64 {
	// score := 0.0
	// for i := 1; i < 4; i++ {
	// 	grams := TokenNgrams(tokens, i)
	// 	sum1 := 0.
	// 	for i := range grams {
	// 		sum1 += o.frequencies.Get(grams[i])
	// 	}
	// 	score += o.weights[i-1] * sum1
	// }

	var score float64

	switch {
	case len(tokens) == 1:
		score += o.calculateUnigramScore(tokens, dist)
	case len(tokens) == 2:
		score += o.calculateBigramScore(tokens, dist)
	case len(tokens) >= 3:
		score += o.calculateTrigramScore(tokens, dist)

	}

	if score == 0 {
		score = math.Inf(-1)
	}
	return score
}
