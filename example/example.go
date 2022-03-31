package main

import (
	"fmt"

	_ "net/http/pprof"

	"github.com/Saimunyz/speller"
)

func main() {
	fmt.Println("Example of usage")
	// create speller
	speller := speller.NewSpeller("../config.yaml")

	// load modelы
	// err := speller.LoadModel("models/sentences.txt.gz")
	// if err != nil {
	// 	fmt.Printf("No such file: %v\n", err)
	// 	//panic(err)
	// }

	// or train model and save
	speller.Train()
	err := speller.SaveModel("models/sentences.txt.gz")
	if err != nil {
		fmt.Printf("No such file: %v\n", err)
		//panic(err)
	}

	// correct typos
	correct := speller.SpellCorrect("канканцелярский")
	fmt.Println("канканцелярский ->", correct)

	correct = speller.SpellCorrect("томат дородгый")
	fmt.Println("томат дородгый ->", correct)
	correct = speller.SpellCorrect("органайзер дородгый")
	fmt.Println("органайзер дородгый ->", correct)

	correct = speller.SpellCorrect("томат дородный")
	fmt.Println("томат дородный ->", correct)

	correct = speller.SpellCorrect("органайзер дорожный")
	fmt.Println("органайзер дорожный ->", correct)

	correct = speller.SpellCorrect("томат дорожный")
	fmt.Println("томат дорожный ->", correct)

	correct = speller.SpellCorrect("органайзер дородный")
	fmt.Println("органайзер дородный ->", correct)

	correct = speller.SpellCorrect("чемодан дородный")
	fmt.Println("чемодан дородный ->", correct)

	correct = speller.SpellCorrect("амоскитная асетка йна впрогулочную акляску")
	fmt.Println("амоскитная асетка йна впрогулочную акляску ->", correct)

	correct = speller.SpellCorrect("брпття сьругацеие ьпудно бытб юоглм")
	fmt.Println("брпття сьругацеие ьпудно бытб юоглм ->", correct)

	correct = speller.SpellCorrect("блузув женчквя серебоистаф")
	fmt.Println("блузув женчквя серебоистаф ->", correct)

	correct = speller.SpellCorrect("обрдрк из воолс в видк кочы")
	fmt.Println("обрдрк из воолс в видк кочы ->", correct)

	correct = speller.SpellCorrect("разршревающее иаслр доя мпчсажа")
	fmt.Println("разршревающее иаслр доя мпчсажа ->", correct)

	correct = speller.SpellCorrect("шиииер ддя воолс")
	fmt.Println("шиииер ддя воолс ->", correct)

	correct = speller.SpellCorrect("раниц жля еачальнфх клвсмов")
	fmt.Println("ранец жля еачальнфх клвсмов ->", correct)
}
