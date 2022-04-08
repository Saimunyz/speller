package main

import (
	"fmt"

	_ "net/http/pprof"

	"github.com/Saimunyz/speller"
)

func main() {
	fmt.Println("Example of usage")
	// create speller
	speller := speller.NewSpeller("config.yaml")

	// load modelы
	err := speller.LoadModel("models/AllRu-model.gz")
	if err != nil {
		fmt.Printf("No such file: %v\n", err)
		//panic(err)
	}

	// or train model and save
	// speller.Train()
	// err := speller.SaveModel("models/AllRu-model.gz")
	// if err != nil {
	// 	fmt.Printf("No such file: %v\n", err)
	// 	//panic(err)
	// }

	// correct typos
	correct := speller.SpellCorrect2("канканцелярский")
	fmt.Println("канканцелярский ->", correct)

	correct = speller.SpellCorrect2("Желтая скатерть")
	fmt.Println("Желтая скатерть ->", correct)

	correct = speller.SpellCorrect2("желиая скаткрть")
	fmt.Println("желиая скаткрть ->", correct)

	correct = speller.SpellCorrect2("томат дородгый")
	fmt.Println("томат дородгый ->", correct)
	correct = speller.SpellCorrect2("органайзер дородгый")
	fmt.Println("органайзер дородгый ->", correct)

	correct = speller.SpellCorrect2("томат дородный")
	fmt.Println("томат дородный ->", correct)

	correct = speller.SpellCorrect2("органайзер дорожный")
	fmt.Println("органайзер дорожный ->", correct)

	correct = speller.SpellCorrect2("томат дорожный")
	fmt.Println("томат дорожный ->", correct)

	correct = speller.SpellCorrect2("органайзер дородный")
	fmt.Println("органайзер дородный ->", correct)

	correct = speller.SpellCorrect2("чемодан дородный")
	fmt.Println("чемодан дородный ->", correct)

	correct = speller.SpellCorrect2("амоскитная асетка йна впрогулочную акляску")
	fmt.Println("амоскитная асетка йна впрогулочную акляску ->", correct)

	correct = speller.SpellCorrect2("брпття сьругацеие ьпудно бытб юоглм")
	fmt.Println("брпття сьругацеие ьпудно бытб юоглм ->", correct)

	correct = speller.SpellCorrect2("блузув женчквя серебоистаф")
	fmt.Println("блузув женчквя серебоистаф ->", correct)

	correct = speller.SpellCorrect2("обрдрк из воолс в видк кочы")
	fmt.Println("обрдрк из воолс в видк кочы ->", correct)

	correct = speller.SpellCorrect2("разршревающее иаслр доя мпчсажа")
	fmt.Println("разршревающее иаслр доя мпчсажа ->", correct)

	correct = speller.SpellCorrect2("шиииер ддя воолс")
	fmt.Println("шиииер ддя воолс ->", correct)

	correct = speller.SpellCorrect2("раниц жля еачальнфх клвсмов")
	fmt.Println("ранец жля еачальнфх клвсмов ->", correct)

	correct = speller.SpellCorrect2("винеи пуз")
	fmt.Println("винеи пуз ->", correct)

	correct = speller.SpellCorrect2("емеость ддя мцсора бкз крфшки")
	fmt.Println("емеость ддя мцсора бкз крфшки ->", correct)

	correct = speller.SpellCorrect2("коем длч куттеулы")
	fmt.Println("коем длч куттеулы ->", correct)

	correct = speller.SpellCorrect2("подмтавкп под пучкт кпаснок аеднрко")
	fmt.Println("подмтавкп под пучкт кпаснок аеднрко ->", correct)
}
