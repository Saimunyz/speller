package main

import (
	"fmt"
	"testing"

	"github.com/Saimunyz/speller"
)

func BenchmarkSpellCheck(b *testing.B) {
	speller1 := speller.NewSpeller("config.yaml")

	// load model
	err := speller1.LoadModel("../models/AllRu-model_new.gz")
	if err != nil {
		fmt.Printf("No such file: %v\n", err)
		panic(err)
	}
	b.ResetTimer()
	b.Run("test", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			speller1.SpellCorrect3("детчкий кпем зажмвай ка от чсадин царааок и ущибов с масласи обоепихи яты и шалыея мое солнышкл")
		}
	})
}
