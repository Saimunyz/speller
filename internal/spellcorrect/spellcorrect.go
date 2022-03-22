package spellcorrect

import (
	"bufio"
	"io"
	"log"
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
			newStr := strings.Builder{}
			newStr.Grow(len(a[i]) + len(b[j]) + 1)
			newStr.WriteString(a[i])
			newStr.WriteString(" ")
			newStr.WriteString(b[j])

			items = append(items, newStr.String())
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
	allSuggestions := make([][]string, len(tokens))

	for i := range tokens {
		// dont look at short words
		if len([]rune(tokens[i])) < 2 {
			allSuggestions[i] = append(allSuggestions[i], tokens[i])
		}

		// gets suggestions
		// o.spell.MaxEditDistance = 2
		suggestions, _ := o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		// if len(suggestions) == 0 {
		// suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		// if len(suggestions) == 0 {
		// o.spell.MaxEditDistance = 3
		// suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelClosest))
		// if len(suggestions) == 0 {
		// 	suggestions, _ = o.spell.Lookup(tokens[i], spell.SuggestionLevel(spell.LevelAll))
		// }
		// }
		// }

		// if we got a word == token and that word's Freq > 50 returns it
		for _, sug := range suggestions {
			if sug.Word == tokens[i] {
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

// getInsertPosition - returns the position sorted in descending order
func getInsertPosition(nums []Suggestion, target Suggestion) int {
	min := 0
	max := len(nums) - 1
	for min <= max {
		mid := min + (max-min)/2
		// temporarily
		if nums[mid].score == 0 {
			nums[mid].score = -1
		}

		switch {
		case target.score == nums[mid].score:
			return mid
		case target.score > nums[mid].score:
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

// getSuggestionCandidates - returns slice of fixed typos with context N-grams
func (o *SpellCorrector) getSuggestionCandidates(allSuggestions [][]string) []Suggestion {
	// combine suggestions
	suggestionStrings := combos(allSuggestions)
	seen := make(map[uint64]struct{}, len(suggestionStrings))
	suggestions := make([]Suggestion, 10)
	for i := range suggestionStrings {
		sugTokens := strings.Split(suggestionStrings[i], " ")
		h := hashTokens(sugTokens)
		if _, ok := seen[h]; !ok {
			seen[h] = struct{}{}
			sugges := Suggestion{
				score:  o.score(sugTokens),
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
