package speller

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Saimunyz/speller/internal/config"
	"github.com/Saimunyz/speller/internal/spellcorrect"
)

type Speller struct {
	spellcorrector *spellcorrect.SpellCorrector
	cfg            *config.Config
}

// NewSpeller - creates new speller instance
func NewSpeller(configPapth string) *Speller {
	cfg, err := config.ReadConfigYML(configPapth)
	if err != nil {
		log.Fatal(err)
	}

	tokenizerWords := spellcorrect.NewSimpleTokenizer()
	freq := spellcorrect.NewFrequencies(cfg.SpellerConfig.MinWordLength, cfg.SpellerConfig.MinWordFreq)

	weights := []float64{
		cfg.SpellerConfig.UnigramWeight,
		cfg.SpellerConfig.BigramWeight,
		cfg.SpellerConfig.TrigramWeight,
	}

	sc := spellcorrect.NewSpellCorrector(
		tokenizerWords,
		freq,
		weights,
		cfg.SpellerConfig.AutoTrainMode,
		cfg.SpellerConfig.MinWordFreq,
		cfg.SpellerConfig.Penalty,
	)

	spller := &Speller{
		spellcorrector: sc,
		cfg:            cfg,
	}
	return spller
}

// Train - train from zero n-grams model with specified in cfg datasets
func (s *Speller) Train() {
	file, err := os.Open(s.cfg.SpellerConfig.SentencesPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	file2, err := os.Open(s.cfg.SpellerConfig.DictPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()

	gz2, err := gzip.NewReader(file2)
	if err != nil {
		log.Fatal(err)
	}
	defer gz2.Close()

	log.Printf("starting training...")
	t0 := time.Now()
	s.spellcorrector.Train(gz, gz2)
	t1 := time.Now()
	log.Printf("Finished[%s]\n", t1.Sub(t0))

	//free memory
	runtime.GC()
}

func (s *Speller) splitByWords(line string, amountOfWords int) []string {
	if len(line) < 1 {
		return []string{}
	}

	words := strings.Fields(line)
	if len(words) <= amountOfWords {
		return []string{line}
	}

	var lines []string

	for i := 0; i+amountOfWords <= len(words); i++ {
		lines = append(lines, strings.Join(words[i:i+amountOfWords], " "))
	}

	// for i := 0; i < len(words); i += amountOfWords {
	// 	stop := i + amountOfWords
	// 	if stop > len(words) {
	// 		start := len(words) - amountOfWords
	// 		lines = append(lines, strings.Join(words[start:], " "))
	// 	} else {
	// 		lines = append(lines, strings.Join(words[i:stop], " "))
	// 	}
	// }

	return lines
}

func (o *Speller) joinByWords(lines []string, splitedByWords int) string {
	if len(lines) < 1 {
		return ""
	}
	words := strings.Fields(lines[0])
	if len(words) < splitedByWords {
		return lines[0]
	}

	query := strings.Builder{}

	for _, line := range lines {
		words = strings.Fields(line)
		query.Grow(len([]rune(words[0])) + 1)
		query.WriteString(words[0])
		query.WriteRune(' ')
	}

	for _, word := range words[1:] {
		query.Grow(len([]rune(word)) + 1)
		query.WriteString(word)
		query.WriteRune(' ')
	}

	return strings.TrimSpace(query.String())
}

func (s *Speller) SpellCorrect2(query string) string {
	if len(query) < 1 {
		return query
	}
	var suggestions []string
	spltQuery := strings.Fields(query)
	shortWords := make(map[int]string) // saves index and short words
	longWords := make([]string, 0, len(spltQuery))
	for i, word := range spltQuery {
		if len([]rune(word)) < s.cfg.SpellerConfig.MinWordLength {
			shortWords[i] = word
			continue
		}
		longWords = append(longWords, word)
	}
	for key, value := range shortWords {
		shortWords[key] = s.spellcorrector.SpellCorrectWithoutContext(value)[0]
	}

	queries := s.splitByWords(strings.Join(longWords, " "), 3)
	for _, query := range queries {
		suggestion := s.spellcorrector.SpellCorrect(query)
		suggestions = append(suggestions, strings.Join(suggestion[0].Tokens, " "))
	}

	joined := s.joinByWords(suggestions, 3)
	words := strings.Fields(joined)

	var extInd int
	var result string
	for j := range spltQuery {
		if word, ok := shortWords[j]; ok {
			result = strings.Join([]string{result, word}, " ")
			continue
		}
		result = strings.Join([]string{result, words[extInd]}, " ")
		extInd++
	}

	// returns the most likely option
	return strings.TrimSpace(result)
}

//SpellCorrect - corrects all typos in a given query
func (s *Speller) SpellCorrect(query string) string {
	if len(query) < 1 {
		return query
	}

	var suggestions []string

	queries := s.splitByWords(query, 3)
	for _, query := range queries {
		suggestion := s.spellcorrector.SpellCorrect(query)
		suggestions = append(suggestions, strings.Join(suggestion[0].Tokens, " "))
	}

	result := s.joinByWords(suggestions, 3)

	// returns the most likely option
	return result
}

// SpellCorrectAllSuggestions - returns top 10 corrections typos in a given query
// func (s *Speller) SpellCorrectAllSuggestions(query string) []string {
// 	suggestions := s.spellcorrector.SpellCorrect(query)

// 	topSugges := make([]string, 0, 10)

// 	for i := 0; i < 10 && i < len(suggestions); i++ {
// 		topSugges = append(topSugges, strings.Join(suggestions[i].Tokens, " "))
// 	}

// 	return topSugges
// }

// SaveModel - saves trained speller model
func (s *Speller) SaveModel(filename string) error {
	fmt.Println("Model saving...")
	err := s.spellcorrector.SaveModel(filename)
	if err != nil {
		return err
	}
	fmt.Printf("Model saved: %s\n", filename)

	return nil
}

// LoadModel - loades trained speller model from file
func (s *Speller) LoadModel(filename string) error {
	t := time.Now()
	fmt.Println("Model loading...")
	err := s.spellcorrector.LoadModel(filename)
	if err != nil {
		return err
	}

	file2, err := os.Open(s.cfg.SpellerConfig.DictPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()

	gz, err := gzip.NewReader(file2)
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	err = s.spellcorrector.LoadFreqDict(gz)
	if err != nil {
		return err
	}
	fmt.Printf("Model loaded[%v]: %s\n", time.Since(t), filename)

	return nil
}
