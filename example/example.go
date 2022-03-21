package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Saimunyz/speller"
)

func main() {
	fmt.Println("Example of usage")
	// create speller
	speller := speller.NewSpeller()

	// load model
	err := speller.LoadModel("../models/test.gz")
	if err != nil {
		fmt.Printf("No such file: %v\n", err)
		//panic(err)
	}

	// or train model and save
	// speller.Train()
	// err := speller.SaveModel("models/test.gz")
	// if err != nil {
	// 	panic(err)
	// }

	// correct typos
	// correct := speller.SpellCorrect("канканцелярский")
	// fmt.Println(correct)

	// correct = speller.SpellCorrect("амоскитная асетка йна впрогулочную акляску")
	// fmt.Println(correct)

	// correct = speller.SpellCorrect("альчик в паласатай аижами всмортить анлаин бесплатно хорошами каче")
	// fmt.Println(correct)

	// correct = speller.SpellCorrect("много блеблет газад мi прелетели йна эьу плонету аосле долгого дляпутишествия ашдгь")
	// fmt.Println(correct)

	correct := speller.SpellCorrect("логитеч клавиатура")
	fmt.Println(correct)

	// p, _ := os.Create("allocs.pprof")
	// defer p.Close()

	// p1, _ := os.Create("heap-after.pprof")
	// defer p1.Close()

	// p2, _ := os.Create("heap-before.pprof")
	// defer p2.Close()

	f, _ := ioutil.ReadFile("test.txt")
	lines := strings.Split(string(f), "\n")
	lines = lines[:len(lines)-1]

	out, _ := os.Create("test-out.txt")

	now := time.Now()

	for _, line := range lines {
		// elem := time.Now()
		// if i == 100000 {
		// 	fmt.Println("stop")
		// }

		out.WriteString(speller.SpellCorrect(line) + "\n")
		// 	// fmt.Println("per elem:", time.Since(elem))
	}
	fmt.Println("time:", time.Since(now))
	// pprof.Lookup("heap").WriteTo(p2, 0)
	runtime.GC()
	// pprof.Lookup("allocs").WriteTo(p, 0)
	// pprof.Lookup("heap").WriteTo(p1, 0)
	fmt.Println("waiting .....")
	// fmt.Scanf("%s")
	time.Sleep(time.Second * 30)

	correct = speller.SpellCorrect("логитеч клавлитура")
	fmt.Println(correct)

}
