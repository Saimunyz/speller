package bagOfWords

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

// bagOfWords - storage of words freq and incoming queriess
type BagOfWords struct {
	sync.Mutex
	Words         map[string]uint64
	Queries       map[string]uint64
	FreqThreshold int
	Expiration    time.Duration
	SendingTime   time.Duration
	OutWords      chan []string
	OutQueries    chan []string
	started       bool
	stop          context.CancelFunc
}

// NewBagOfWords - returns new BagOfWords instance
func NewBagOfWords(cleaningTime time.Duration, sendingTime time.Duration, freqThreshold int, stop context.CancelFunc) *BagOfWords {
	bag := &BagOfWords{
		Words:         make(map[string]uint64),
		Queries:       make(map[string]uint64),
		FreqThreshold: freqThreshold,
		Expiration:    cleaningTime,
		SendingTime:   sendingTime,
		OutWords:      make(chan []string, 1000),
		OutQueries:    make(chan []string, 1000),
		stop:          stop,
	}

	return bag
}

func (b *BagOfWords) isReady() bool {
	b.Lock()
	defer b.Unlock()

	return b.started
}

// Start - starting all cycles in bag
func (b *BagOfWords) Start(ctx context.Context) {
	// starting expiration time
	go b.Clean(ctx)
	go b.Send(ctx)

	b.Lock()
	defer b.Unlock()
	b.started = true
}

// Stop - stopping all cycles in bag
func (b *BagOfWords) Stop() {
	b.stop()
	b.Lock()
	defer b.Unlock()
	b.started = false

}

// Send - sending data to traning in SendingTime
func (b *BagOfWords) Send(ctx context.Context) {
	log.Println("Sending data from bag of words started")
	for {
		select {
		case <-ctx.Done():
			close(b.OutQueries)
			close(b.OutWords)
			log.Println("Sending stopped")
			return
		case <-time.After(b.SendingTime):
			// sending data before cleaning
			_, queries := b.getAllCandidates()
			// if len(words) != 0 || len(queries) != 0 {
			// 	log.Printf("data sending:\n Words: %v \nQueries: %v\n", words, queries)
			// }

			// sending words
			// b.OutWords <- words

			// sending queries
			if len(queries) != 0 {
				b.OutQueries <- queries
			}
		}

	}
}

// Clean - clean bagOfWords in Expiration tima
func (b *BagOfWords) Clean(ctx context.Context) {
	log.Println("Cleaning bag started")
	for {
		select {
		case <-ctx.Done():
			log.Println("Cleaning stopped")
			return
		case <-time.After(b.Expiration):
			if b.Words == nil || b.Queries == nil {
				log.Fatal("Error: Map has nil value")
			}

			// sending data before cleaning
			_, queries := b.getAllCandidates()
			// if len(words) != 0 || len(queries) != 0 {
			// 	log.Printf("data sending before cleaning:\n Words: %v \nQueries: %v\n", words, queries)
			// }

			// sending words
			// b.OutWords <- words

			// sending queries
			if len(queries) != 0 {
				b.OutQueries <- queries
			}

			// clean bag of words
			b.Lock()
			if len(b.Words) != 0 {
				b.Words = make(map[string]uint64)
			}
			if len(b.Queries) != 0 {
				b.Queries = make(map[string]uint64)
			}
			b.Unlock()
			log.Println("Bag of words is cleared")
		}
	}
}

// Add - adds new elements in bag of words
func (b *BagOfWords) Add(query string) {
	if !b.isReady() {
		return
	}
	b.Lock()
	defer b.Unlock()

	log.Println("Addning new query to the bag", query)

	query = strings.ToLower(query)
	words := strings.Fields(query)

	b.Queries[query]++

	for _, word := range words {
		b.Words[word]++
	}
}

// repeat - repeats N times query and returns slice of it
func (b *BagOfWords) repeat(query string, times uint64) []string {
	result := make([]string, times)
	for i := uint64(0); i < times; i++ {
		result[i] = query
	}

	return result
}

// getWordsCandidates - returns all words that greater than threshold
func (b *BagOfWords) getWordsCandidates() (words []string) {
	b.Lock()
	defer b.Unlock()

	for word, freq := range b.Words {
		if freq >= uint64(b.FreqThreshold) {
			words = append(words, b.repeat(word, freq)...)

			// delete this word
			delete(b.Words, word)
		}
	}

	return words
}

// getQueriesCandidates - returns all queries that greater than threshold
func (b *BagOfWords) getQueriesCandidates() (queries []string) {
	b.Lock()
	defer b.Unlock()

	for query, freq := range b.Queries {
		if freq >= uint64(b.FreqThreshold) {
			queries = append(queries, b.repeat(query, freq)...)

			// delete this query
			delete(b.Queries, query)
		}
	}

	return queries
}

// getAllCandidates - returns all words and queries that greater than threshold
func (b *BagOfWords) getAllCandidates() (words []string, queries []string) {
	words = b.getWordsCandidates()
	queries = b.getQueriesCandidates()

	return words, queries
}
