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
	"unicode"

	"github.com/eskriett/spell"
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
	TrainNgramsOnline(tokens []string) error
}

// Tokinizer - tokenizer function from token layer
type Tokenizer interface {
	Tokens(in io.Reader) ([]string, error)
}

type SpellCorrector struct {
	tokenizer     Tokenizer
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
		tokenizer:     tokenizer,
		frequencies:   frequencies,
		spell:         spell.New(),
		weights:       weights,
		minFreq:       minFreq,
		penalty:       penalty,
		autoTrainMode: autoTrainMode,
	}
	ans.spell.MaxEditDistance = 3
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
func (o *SpellCorrector) LoadFreqDict(in io.Reader) error {
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
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Train - train n-grams model from in Reader and read freq from in2 Reader
func (o *SpellCorrector) Train(in io.Reader, in2 io.Reader) error {
	// counting n-grams freq
	err := o.frequencies.TrainNgrams(in)
	if err != nil {
		return err
	}
	runtime.GC()

	// load freq dict
	t := time.Now()
	err = o.LoadFreqDict(in2)
	if err != nil {
		return err
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
func product(a []string, b []string) []string {
	size := len(a) * len(b)
	items := make([]string, 0, size)
	for i := range a {
		for j := range b {
			items = append(items, a[i]+" "+b[j])
		}
	}
	return items
}

// combos - permutation of all sentences
func combos(in [][]string) []string {
	tmpP := in[len(in)-1]
	for i := len(in) - 2; i >= 0; i-- {
		tmpP = product(in[i], tmpP)
	}
	return tmpP
}

// lookupTokens - finds all the suggestions given by the spell library and takes the top 20 of them
func (o *SpellCorrector) lookupTokens(tokens []string) ([][]string, map[string]float64) {
	allSuggestions := make([][]string, len(tokens))
	dist := make(map[string]float64)

	for i := range tokens {
		// dont look at short words
		if len([]rune(tokens[i])) < 3 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
			dist[tokens[i]] = 0
		}

		// gets suggestions
		var suggestions spell.SuggestionList
		o.spell.MaxEditDistance = 2

		suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		if len(suggestions) == 0 {
			o.spell.MaxEditDistance = 3
			suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		}

		// gets 5 first suggestions
		if len(allSuggestions[i]) == 0 {
			for j := 0; j < len(suggestions) && j < 5; j++ {
				allSuggestions[i] = append(allSuggestions[i], suggestions[j].Word)
				dist[suggestions[j].Word] = float64(suggestions[j].Distance) + float64(j)*o.penalty
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
		// case target.score == nums[mid].score:
		// 	return mid
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
		sugTokens := strings.Split(suggestionStrings[i], " ")
		h := hashTokens(sugTokens)
		if _, ok := seen[h]; !ok {
			seen[h] = struct{}{}
			sugges := Suggestion{
				score:  o.score(sugTokens, dist),
				Tokens: sugTokens,
			}
			pos := getInsertPosition(suggestions, sugges)
			insertPosition(suggestions, pos, sugges)
		}
	}

	return suggestions
}

func (o *SpellCorrector) addWordToModel(newWords chan string) {
	for query := range newWords {
		var tokens []string

		words := strings.Fields(query)
		for _, word := range words {
			if len([]rune(word)) < 2 {
				continue
			}
			word = strings.TrimRightFunc(word, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			word = strings.ToLower(word)
			tokens = append(tokens, word)

			// update spell library
			entry, err := o.spell.GetEntry(word)
			if err != nil || entry == nil {
				// add new entry
				o.spell.AddEntry(spell.Entry{
					Frequency: 1,
					Word:      word,
				})
				continue
			}
			o.spell.AddEntry(spell.Entry{
				Frequency: entry.Frequency + 1,
				Word:      entry.Word,
			})
		}
		o.frequencies.TrainNgramsOnline(tokens)
	}
}

// SpellCorrect - returns suggestions
func (o *SpellCorrector) SpellCorrect(s string) []Suggestion {
	// new words for model improvments
	newWords := make(chan string)
	if o.autoTrainMode {
		go o.addWordToModel(newWords)
	}

	tokens, _ := o.tokenizer.Tokens(strings.NewReader(s))
	allSuggestions, dist := o.lookupTokens(tokens)
	items := o.getSuggestionCandidates(allSuggestions, dist)

	// sending data to model improvments
	if o.autoTrainMode {
		go func() {
			sugges := strings.Join(items[0].Tokens, " ")
			newWords <- sugges
			newWords <- s
		}()
	}

	return items
}

func (o *SpellCorrector) SpellCorrectWithoutContext(s string) []string {
	if len([]rune(s)) < 3 {
		return []string{s}
	}

	suggestions, _ := o.spell.Lookup(s, spell.SuggestionLevel(spell.LevelClosest))
	result := make([]string, len(suggestions))

	for i := range result {
		result[i] = suggestions[i].Word
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
	scaled_value := 1 + (10-1)*relative

	alpha = math.Log10(scaled_value) * 100

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
	var (
		uniLog float64
		biLog  float64
		score  float64
	)

	// penalty := len(bigrams)

	for i := 0; i+2 <= len(ngrams); i++ {
		bigrams := ngrams[i : i+2 : i+2]

		bigram := o.frequencies.Get(bigrams)
		if bigram != 0 {
			biLog = math.Log(bigram)
			biLog -= getPenalty(biLog, dist[bigrams[0]]+dist[bigrams[1]])

			unigram := o.GetUnigram(bigrams)
			if unigram != 0 {
				uniLog = math.Log(unigram) + o.weights[0]
				uniLog -= getPenalty(uniLog, dist[bigrams[0]])
			}

			score += uniLog + biLog
		} else {
			tmp := o.calculateUnigramScore(bigrams, dist)
			score += (tmp + tmp)
		}
	}
	return score
}

// calculateUnigramScore - returns unigram score of a given words
func (o *SpellCorrector) calculateUnigramScore(ngrams []string, dist map[string]float64) float64 {
	var (
		uniLog float64
		score  float64
	)

	penalty := len(ngrams)

	for i := 0; i+1 <= len(ngrams); i++ {
		unigrams := ngrams[i : i+1 : i+1]

		unigram := o.frequencies.Get(unigrams)
		if unigram != 0 {
			penalty--
			uniLog = math.Log(unigram)
			uniLog -= getPenalty(uniLog, dist[unigrams[0]])
		}

		score += uniLog
	}

	if penalty > 0 {
		score = o.applyPenalty(score, penalty)
	}
	return score
}

// calculateTrigramScore -  returns trigrams score of a given words
func (o *SpellCorrector) calculateTrigramScore(ngrams []string, dist map[string]float64) float64 {
	var (
		uniLog float64
		biLog  float64
		triLog float64
		score  float64
	)

	for i := 0; i+3 <= len(ngrams); i++ {
		trigrams := ngrams[i : i+3 : i+3]

		trigram := o.frequencies.Get(trigrams)
		if trigram != 0 {
			triLog = math.Log(trigram)
			triLog -= getPenalty(triLog, dist[trigrams[0]]+dist[trigrams[1]]+dist[trigrams[2]])

			bigram := o.GetBigram(trigrams)
			if bigram != 0 {
				biLog = math.Log(bigram) + o.weights[1]
				biLog -= getPenalty(biLog, dist[trigrams[0]]+dist[trigrams[1]])
			}
			unigram := o.GetUnigram(trigrams)
			if unigram != 0 {
				uniLog = math.Log(unigram) + o.weights[0]
				uniLog -= getPenalty(uniLog, dist[trigrams[0]])
			}

			score += uniLog + biLog + triLog
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

	// ngrams := TokenNgrams(tokens, 3)
	// if len(ngrams) == 0 {
	// 	ngrams = TokenNgrams(tokens, 2)
	// 	if len(ngrams) == 0 {
	// 		ngrams = TokenNgrams(tokens, 1)
	// 	}
	// }

	// if len(tokens) > 2 {
	// 	if tokens[0] == "орел" && tokens[1] == "заправляет" && tokens[2] == "крылья" {
	// 		fmt.Println(score)
	// 	}
	// 	if tokens[0] == "орел" && tokens[1] == "расправляет" && tokens[2] == "крылья" {
	// 		fmt.Println(score)
	// 	}
	// }

	// for i := range ngrams {
	// 	switch len(ngrams[i]) {
	// 	case 1:
	// 		score += o.calculateUnigramScore(ngrams[i], dist)
	// 	case 2:
	// 		score += o.calculateBigramScore(ngrams[i], dist)
	// 	case 3:
	// 		score += o.calculateTrigramScore(ngrams[i], dist)
	// 	}
	// }
	// if score != 0 {
	// 	score = math.Exp(score)
	// }
	if score == 0 {
		score = math.Inf(-1)
	}
	return score
}
