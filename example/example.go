package main

import (
	"fmt"

	"github.com/Saimunyz/speller"
)

func main() {
	fmt.Println("Example of usage")
	// create speller
	speller := speller.NewSpeller()

	// load model
	err := speller.LoadModel("models/bin-not-normalized-data.gz")
	if err != nil {
		fmt.Printf("No such file: %v\n", err)
		//panic(err)
	}

	// or train model and save
	// speller.Train()
	// err := speller.SaveModel("models/bin-not-normalized-data.gz")
	// if err != nil {
	// 	panic(err)
	// }

	// correct typos
	correct := speller.SpellCorrect("концелярсикй")
	fmt.Println(correct)

	correct = speller.SpellCorrect("логитеч клавиатура")
	fmt.Println(correct)

	correct = speller.SpellCorrect("логитеч клавиатура")
	fmt.Println(correct)

	correct = speller.SpellCorrect("логитеч клавиатура")
	fmt.Println(correct)

	correct = speller.SpellCorrect("логитеч клавиатура")
	fmt.Println(correct)
}
