package bagOfWords

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestGetCandidates(t *testing.T) {
	data := []string{
		"первый запрос",
		"клавиатура логитеч",
		"клавиатура логитеч",
		"клавиатура логитеч",
		"клавиатура логитеч",
		"клавиатура логитеч",
	}

	cleaningTime := time.Second * 10
	sendingTime := time.Second * 5
	freqThreshold := 5
	ctx, cancel := context.WithCancel(context.Background())

	bag := NewBagOfWords(
		cleaningTime,
		sendingTime,
		freqThreshold,
		cancel,
	)

	bag.Start(ctx)

	for _, query := range data {
		bag.Add(query)
	}

	bag.Stop()

	if len(bag.Words) != 4 {
		t.Error("Not all words was added")
		os.Exit(1)
	}

	if len(bag.Queries) != 2 {
		t.Error("Not all queries was added")
		os.Exit(1)
	}

	words, queries := bag.getAllCandidates()
	if len(words) != 10 || len(queries) != 5 {
		t.Error("Wrong amout of words or queries from getAllCandidates")
		os.Exit(1)
	}
}

func TestRepeat(t *testing.T) {
	cleaningTime := time.Second * 10
	sendingTime := time.Second * 5
	freqThreshold := 5
	ctx, cancel := context.WithCancel(context.Background())

	bag := NewBagOfWords(
		cleaningTime,
		sendingTime,
		freqThreshold,
		cancel,
	)
	bag.Start(ctx)
	bag.Stop()

	expected := []string{
		"hello",
		"hello",
		"hello",
	}

	result := bag.repeat("hello", 3)

	if len(result) != len(expected) {
		t.Error("Repeat not same length")
		os.Exit(1)
	}

	for i := 0; i < len(result); i++ {
		if result[i] != expected[i] {
			t.Error("result and expexted not equal", "result:", result[i], "expext:", expected[i])
			os.Exit(1)
		}
	}

}

func TestClean(t *testing.T) {
	cleaningTime := time.Second * 1
	sendingTime := time.Second * 30
	freqThreshold := 5
	ctx, cancel := context.WithCancel(context.Background())

	bag := NewBagOfWords(
		cleaningTime,
		sendingTime,
		freqThreshold,
		cancel,
	)
	bag.Start(ctx)

	bag.Add("Hello World!")

	time.Sleep(time.Second * 2)
	bag.Stop()

	if len(bag.Words) != 0 {
		t.Errorf("bag words wasn't cleared")
		os.Exit(1)
	}

	if len(bag.Queries) != 0 {
		t.Errorf("bag queries wasn't cleared")
		os.Exit(1)
	}
	cancel()
}

func TestSend(t *testing.T) {
	cleaningTime := time.Second * 50
	sendingTime := time.Second * 1
	freqThreshold := 1
	ctx, cancel := context.WithCancel(context.Background())

	bag := NewBagOfWords(
		cleaningTime,
		sendingTime,
		freqThreshold,
		cancel,
	)
	bag.Start(ctx)

	bag.Add("Hello World!")

	time.Sleep(time.Second * 2)
	bag.Stop()

	totalWords := make([]string, 0, 2)
	totalQueries := make([]string, 0, 1)
	for word := range bag.OutWords {
		totalWords = append(totalWords, word...)
	}
	for query := range bag.OutQueries {
		totalQueries = append(totalQueries, query...)
	}

	if len(totalWords) != 2 {
		t.Errorf("bag words chan isn't worked")
		os.Exit(1)
	}

	if len(totalQueries) != 1 {
		t.Errorf("bag query chan isn't worked")
		os.Exit(1)
	}

}
