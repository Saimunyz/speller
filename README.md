# speller
Spelling correction using ngrams and symspell

To work correctly, you need to have config.yaml locally and files from dataset folder

```
	fmt.Println("Example of usage")
	// create speller
	speller := speller.NewSpeller()

	// load model
	err := speller.LoadModel("models/small-data")
	if err != nil {
		fmt.Printf("No such file: %v\n", err)
		//panic(err)
	}

	// // or train model and save
	// speller.Train()
	// err = speller.SaveModel("models/small-data")
	// if err != nil {
	// 	panic(err)
	// }

	// correct typos
	correct := speller.SpellCorrect("концелярсикй")
	fmt.Println(correct)

```

