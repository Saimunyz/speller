package spellcorrect

import (
	"bufio"
	"io"
	"log"
	"runtime"
	"sort"
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
	autoTrainMode bool
}

// NewSpellCorrector - creates new SpellCorrector instance
func NewSpellCorrector(tokenizer Tokenizer, frequencies FrequencyContainer, weights []float64, autoTrainMode bool) *SpellCorrector {
	ans := SpellCorrector{
		tokenizer:     tokenizer,
		frequencies:   frequencies,
		spell:         spell.New(),
		weights:       weights,
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
func (o *SpellCorrector) lookupTokens(tokens []string) [][]string {
	allSuggestions := make([][]string, 0, len(tokens))
	for i := range tokens {
		allSuggestions = append(allSuggestions, nil) // why so?
		// dont look at short words
		if len([]rune(tokens[i])) < 2 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
		}

		// gets suggestions
		var sugges []spell.Suggestion
		suggestions, _ := o.spell.Lookup(tokens[i], spell.SuggestionLevel(5))
		// takes only with distance of 2 and lower
		for _, sug := range suggestions {
			if sug.Distance < 3 {
				sugges = append(sugges, sug)
			}
		}

		// if there is no offer from distance 2 receives from 3
		if len(sugges) != 0 {
			suggestions = sugges
		}

		// if we got a word == token and that word's Freq > 50 returns it
		for _, sug := range suggestions {
			if sug.Word == tokens[i] && sug.Frequency >= 50 {
				allSuggestions[i] = append(allSuggestions[i], tokens[i])
				break
			}
		}

		// if no words == token gets 20 first suggestions
		if len(allSuggestions[i]) == 0 {
			for j := 0; j < len(suggestions) && j < 20; j++ {
				allSuggestions[i] = append(allSuggestions[i], suggestions[j].Word)
			}
		}

		// if no suggestions returns token
		if len(allSuggestions[i]) == 0 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
		}
	}
	return allSuggestions
}

// getSuggestionCandidates - returns slice of fixed typos with context N-grams
func (o *SpellCorrector) getSuggestionCandidates(allSuggestions [][]string) []Suggestion {
	// combine suggestions
	suggestionStrings := combos(allSuggestions)
	seen := make(map[uint64]struct{}, len(suggestionStrings))
	suggestions := make([]Suggestion, 0, len(suggestionStrings))
	for i := range suggestionStrings {
		sugTokens := strings.Split(suggestionStrings[i], " ")
		h := hashTokens(sugTokens)
		if _, ok := seen[h]; !ok {
			seen[h] = struct{}{}
			suggestions = append(suggestions,
				Suggestion{
					// Score each word/sentence
					score:  o.score(sugTokens),
					Tokens: sugTokens,
				})
		}
	}
	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].score > suggestions[j].score
	})
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
	allSuggestions := o.lookupTokens(tokens)
	items := o.getSuggestionCandidates(allSuggestions)

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

// scpre - scoring each sentence
func (o *SpellCorrector) score(tokens []string) float64 {
	score := 0.0
	for i := 1; i < 4; i++ {
		grams := TokenNgrams(tokens, i)
		sum1 := 0.
		for i := range grams {
			sum1 += o.frequencies.Get(grams[i])
		}
		score += o.weights[i-1] * sum1
	}
	return score
}