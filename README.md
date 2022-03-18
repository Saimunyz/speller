# speller
Spelling correction using ngrams and symspell

To work correctly, you need to have config.yaml and two files:
file freq-dict.txt.gz - gzip text file with content:
word freq
example:
```
hello 124
words 222
```

file sentences.txt.gz - gzip txt file with big text fro traning ngrams model

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

