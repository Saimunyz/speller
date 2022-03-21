package speller

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Saimunyz/speller/internal/bagOfWords"
	"github.com/Saimunyz/speller/internal/config"
	"github.com/Saimunyz/speller/internal/spellcorrect"
)

type Speller struct {
	spellcorrector *spellcorrect.SpellCorrector
	cfg            *config.Config
	ready          chan struct{}
}

// NewSpeller - creates new speller instance
func NewSpeller() *Speller {
	cfg, err := config.ReadConfigYML("../config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	tokenizerWords := spellcorrect.NewSimpleTokenizer()
	freq := spellcorrect.NewFrequencies(cfg.SpellerConfig.MinWordLength, cfg.SpellerConfig.MinWordFreq)

	ctx, cancel := context.WithCancel(context.Background())

	bagOfWords := bagOfWords.NewBagOfWords(
		time.Minute*time.Duration(cfg.SpellerConfig.CleaningTime),
		time.Second,
		cfg.SpellerConfig.FreqThreshold,
		cancel,
	)

	weights := []float64{cfg.SpellerConfig.UnigramWeight, cfg.SpellerConfig.BigramWeight, cfg.SpellerConfig.TrigramWeight}
	sc := spellcorrect.NewSpellCorrector(tokenizerWords, freq, weights, bagOfWords, cancel)

	spller := &Speller{
		spellcorrector: sc,
		cfg:            cfg,
		ready:          make(chan struct{}),
	}

	if cfg.SpellerConfig.AutoTrainMode {
		go func() {
			<-spller.ready
			bagOfWords.Start(ctx)
			sc.StartAutoTrain()
		}()
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

	// ready signal
	// s.ready <- struct{}{}
}

//SpellCorrect - corrects all typos in a given query
func (s *Speller) SpellCorrect(query string) string {
	var result string

	// splitting query by 4 words lenght
	words := strings.Fields(query)
	if len(words) > 4 {
		var shortQueries []string
		for i := 0; i < len(words); i += 4 {
			stop := i + 4
			if i+4 >= len(words) {
				stop = len(words)
			}
			shortQuery := strings.Join(words[i:stop:stop], " ")
			suggestion := s.spellcorrector.SpellCorrect(shortQuery)
			shortQueries = append(shortQueries, suggestion[0].Tokens...)
		}
		result = strings.Join(shortQueries, " ")
	} else {
		suggestions := s.spellcorrector.SpellCorrect(query)
		result = strings.Join(suggestions[0].Tokens, " ")
	}

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

	file2, err := os.Open("../" + s.cfg.SpellerConfig.DictPath)
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

	// ready signal
	s.ready <- struct{}{}

	return nil
}
