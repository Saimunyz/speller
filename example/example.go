package main

import (
	"fmt"
	"time"

	_ "net/http/pprof"

	"github.com/Saimunyz/speller"
)

func main() {
	fmt.Println("Example of usage")
	// create speller
	speller := speller.NewSpeller("config.yaml")

	// load model
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

	now := time.Now()

	// correct typos
	// correct := speller.SpellCorrect3("канканцелярский")
	// fmt.Println("канканцелярский ->", correct)

	// correct = speller.SpellCorrect3("жеский")
	// fmt.Println("жеское ->", correct)

	correct := speller.SpellCorrect("кружеа иреуголка л бедо кормчневая тиго ьольшая подпрок для мкжчины для жннщины")
	fmt.Println("кружеа иреуголка л бедо кормчневая тиго ьольшая подпрок для мкжчины для жннщины ->", correct)

	correct = speller.SpellCorrect("тайсеая сазь для лечениф гриюка ногтнй гртбок")
	fmt.Println("тайсеая сазь для лечениф гриюка ногтнй гртбок ->", correct)

	correct = speller.SpellCorrect("платя дя женщин")
	fmt.Println("платя дя женщин ->", correct)

	correct = speller.SpellCorrect("красавки")
	fmt.Println("красавки ->", correct)

	correct = speller.SpellCorrect("Желтая скатерть")
	fmt.Println("Желтая скатерть ->", correct)

	correct = speller.SpellCorrect("желиая скаткрть")
	fmt.Println("желиая скаткрть ->", correct)

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

	correct = speller.SpellCorrect("амоскитная асетка на впрогулочную акляску")
	fmt.Println("амоскитная асетка на впрогулочную акляску ->", correct)

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

	correct = speller.SpellCorrect("винеи пуз")
	fmt.Println("винеи пуз ->", correct)

	correct = speller.SpellCorrect("емеость ддя мцсора бкз крфшки")
	fmt.Println("емеость ддя мцсора бкз крфшки ->", correct)

	correct = speller.SpellCorrect("коем длч куттеулы")
	fmt.Println("коем длч куттеулы ->", correct)

	correct = speller.SpellCorrect("цаоь фндрр оркл рачправлфет крыоьч")
	fmt.Println("цаоь фндрр оркл рачправлфет крыоьч ->", correct)

	correct = speller.SpellCorrect("подмтавкп под пучкт кпаснок аеднрко")
	fmt.Println("подмтавкп под пучкт кпаснок аеднрко ->", correct)

	fmt.Println("-----------------")

	correct = speller.SpellCorrect("томат дорожный")
	fmt.Println("томат дорожный ->", correct)

	correct = speller.SpellCorrect("чемодан дородный")
	fmt.Println("чемодан дородный ->", correct)

	correct = speller.SpellCorrect("жесткие толстовки")
	fmt.Println("жесткие толстовки ->", correct)

	correct = speller.SpellCorrect("костюм для тонировок")
	fmt.Println("костюм для тонировок ->", correct)

	correct = speller.SpellCorrect("летний плач")
	fmt.Println("летний плач ->", correct)

	correct = speller.SpellCorrect("набор свечей столик")
	fmt.Println("набор свечей столик ->", correct)

	correct = speller.SpellCorrect("жирное мыло")
	fmt.Println("жирное мыло ->", correct)

	correct = speller.SpellCorrect("фотошторы для пальчика")
	fmt.Println("фотошторы для пальчика ->", correct)

	correct = speller.SpellCorrect("повседневное боги")
	fmt.Println("повседневное боги ->", correct)

	correct = speller.SpellCorrect("подушка разовая")
	fmt.Println("подушка разовая ->", correct)

	fmt.Println(time.Since(now))
}
