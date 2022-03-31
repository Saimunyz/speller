package main

import (
	"github.com/Saimunyz/speller/internal/tokenizer/normalize"
	"log"
	"time"
)

func main() {
	tokenizer := normalize.NewNormalizer()
	t := time.Now().UTC()
	err := tokenizer.LoadDictionariesLocal("./data/words.csv.gz", "./data/spellcheck12.01.csv")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(time.Now().Sub(t))
	log.Println(tokenizer.NormalizeWithoutMeta("синий пудровый"))
	log.Println(tokenizer.NormalizeWithoutMeta("тест/тест2/тест3"))
}
