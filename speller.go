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

const (
	shortWordsDict = "shortWords"
	defaultDict    = "default"
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
	sentencesFile, err := os.Open(s.cfg.SpellerConfig.SentencesPath)
	if err != nil {
		log.Fatal(err)
	}
	defer sentencesFile.Close()

	sentencesGz, err := gzip.NewReader(sentencesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer sentencesGz.Close()

	dictFile, err := os.Open(s.cfg.SpellerConfig.DictPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dictFile.Close()

	dictGz, err := gzip.NewReader(dictFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dictGz.Close()

	commonDict := spellcorrect.FreqDicts{
		Name:   "defatult",
		Reader: dictGz,
	}

	shortWordsFile, err := os.Open(s.cfg.SpellerConfig.ShortWordsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer shortWordsFile.Close()

	shortWordsGz, err := gzip.NewReader(shortWordsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer shortWordsGz.Close()

	shortWordsDict := spellcorrect.FreqDicts{
		Name:   "shortWords",
		Reader: shortWordsGz,
	}

	log.Printf("starting training...")
	t0 := time.Now()
	s.spellcorrector.Train(sentencesGz, commonDict, shortWordsDict)
	t1 := time.Now()
	log.Printf("Finished[%s]\n", t1.Sub(t0))

	//free memory
	runtime.GC()
}

func (s *Speller) splitByWords(words []string, amountOfWords int) [][]string {
	if len(words) < 1 {
		return [][]string{}
	}

	if len(words) <= amountOfWords {
		return [][]string{words}
	}

	lines := make([][]string, len(words)-amountOfWords+1)

	for i := 0; i+amountOfWords <= len(words); i++ {
		lines[i] = words[i : i+amountOfWords]
	}

	return lines
}

func (o *Speller) joinByWords(lines [][]string, splitedByWords int) []string {
	if len(lines) < 1 {
		return []string{}
	}

	if len(lines[0]) < splitedByWords {
		return lines[0]
	}

	query := make([]string, len(lines)+len(lines[len(lines)-1][1:]))

	for i, words := range lines {
		query[i] = words[0]
		if i == len(lines)-1 {
			i++
			for _, word := range words[1:] {
				query[i] = word
				i++
			}
		}
	}

	return query
}

func (s *Speller) SpellCorrect2(query string) string {
	if len(query) < 1 {
		return query
	}

	spltQuery, _ := s.spellcorrector.Tokenizer.Tokens(strings.NewReader(query))
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

	queries := s.splitByWords(longWords, 3)
	for i, query := range queries {
		suggestion := s.spellcorrector.SpellCorrect(query)
		queries[i] = suggestion[0].Tokens
	}

	words := s.joinByWords(queries, 3)

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

	dictFile, err := os.Open(s.cfg.SpellerConfig.DictPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dictFile.Close()

	dictGz, err := gzip.NewReader(dictFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dictGz.Close()

	shortWordsFile, err := os.Open(s.cfg.SpellerConfig.ShortWordsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer shortWordsFile.Close()

	shortWordsGz, err := gzip.NewReader(shortWordsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer shortWordsGz.Close()

	err = s.spellcorrector.LoadFreqDict(dictGz, defaultDict)
	if err != nil {
		return err
	}

	err = s.spellcorrector.LoadFreqDict(shortWordsGz, shortWordsDict)
	if err != nil {
		return err
	}
	fmt.Printf("Model loaded[%v]: %s\n", time.Since(t), filename)

	return nil
}

func (s *Speller) SpellCorrect3(query string) string {
	if len(query) < 1 {
		return query
	}
	wordsToCorrect := make(map[string]struct{})
	spltQuery, _ := s.spellcorrector.Tokenizer.Tokens(strings.NewReader(query))
	shortWords := make(map[int]string)             // храним короткое слово и его индекс
	longWords := make([]string, 0, len(spltQuery)) // длинные слова
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
	//если слово есть в словаре, его не исправляем
	for _, v := range longWords {
		if !s.spellcorrector.CheckInFreqDict(v) {
			wordsToCorrect[v] = struct{}{}
		}
	}
	if len(wordsToCorrect) == 0 { //если нет длинных слов для исправления, то собираем ответ и выходим
		var extInd int
		var result string
		for j := range spltQuery {
			if word, ok := shortWords[j]; ok {
				result = strings.Join([]string{result, word}, " ")
				continue
			}
			result = strings.Join([]string{result, longWords[extInd]}, " ")
			extInd++
		}
		// returns the most likely option
		return strings.TrimSpace(result)
	}

	queries := s.splitByWords(longWords, 3) //генерим триграммы из длинных слов
	for i, query := range queries {
		//если первого слова триграммы нет в словаре, то мы отдаем спеллеру
		if ok := needToFix(query[0], wordsToCorrect); !ok && i != len(queries)-1 {
			//если первое слово триграммы есть в словаре, то мы всю триграмму без изменений добавляем в саджесты
			//потому что при сборке ответа из саджестов, берется только первое слово саджеста
			// suggestions = append(suggestions, query)
			continue
		}
		suggestion := s.spellcorrector.SpellCorrect2(query)
		queries[i] = suggestion[0].Tokens
	}

	words := s.joinByWords(queries, 3) //собрали длинные исправленные слова из саджестов

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

func needToFix(query string, wordsToCorrect map[string]struct{}) bool {
	lastIndxFirstWord := strings.Index(query, " ")
	if lastIndxFirstWord == -1 {
		lastIndxFirstWord = len(query)
	}
	if _, ok := wordsToCorrect[query[:lastIndxFirstWord]]; ok {
		return true
	}
	return false
}
