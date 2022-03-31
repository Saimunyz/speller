package normalize

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
)

type ElasticNormRequest struct {
	Tokenizer 	string 		`json:"tokenizer"`
	Filter 		[]string	`json:"filter"`
	Char_filter []string	`json:"char_filter"`
	Text	 	string		`json:"text"`
}

type ElasticNormResponse struct {
	Tokens []ElasticTokens  `json:"tokens"`
}

type ElasticTokens struct {
	Token 			string	`json:"token"`
	Start_offset 	int		`json:"start_offset"`
	Snd_offset 		int 	`json:"end_offset"`
	Type_ 			string 	`json:"type_"`
	Position 		int		`json:"position"`
}

func BenchmarkNormalizer_Normalize(b *testing.B) {
	tokenizer := NewNormalizer()
	err := tokenizer.LoadDictionariesLocal("../data/words.csv", "../spellcheck/spellcheck12.01.csv")
	if err != nil {
		log.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenizer.Normalize("синии кросовки крутые")
	}
}

func TestNormalizer_Normalize_1(t *testing.T) {
	tokenizer := NewNormalizer()
	err := tokenizer.LoadDictionariesLocal("../data/words.csv_latest1.gz", "../spellcheck/spellcheck12.01.csv")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tokenizer.normalize("пипл-метров"))
	//return

	fileCompositeWords, _ := os.OpenFile("words-composite.csv", 0777, fs.FileMode(os.O_CREATE))
	defer fileCompositeWords.Close()
	fileCompositeWords.WriteString("0\n")

	fileNoCompositeWords, _ := os.OpenFile("words-normal.csv", 0777, fs.FileMode(os.O_CREATE))
	defer fileNoCompositeWords.Close()
	fileNoCompositeWords.WriteString("0\n")

	fileNotRuWords, _ := os.OpenFile("words-notru.csv", 0777, fs.FileMode(os.O_CREATE))
	defer fileNotRuWords.Close()
	fileNotRuWords.WriteString("0\n")

	fileZeroWords, _ := os.OpenFile("words-zero.csv", 0777, fs.FileMode(os.O_CREATE))
	defer fileZeroWords.Close()
	fileZeroWords.WriteString("0\n")

	res1 := make(map[string]string) 
	var resWord [][]LemmaWithMeta
	var someWords []string
	counterComposite := 0
	for k, _ := range tokenizer.words {
		resWord = tokenizer.Normalize(k)
		if len(resWord) == 1 {
			if len(resWord[0]) == 2 {
				if isRussian(resWord[0][1].Lemma) {
					res1[k] = resWord[0][1].Lemma
					//fileNoCompositeWords.WriteString(fmt.Sprintf("%s;%s;%s;%d\n", k, v.Lemma, v.POS, v.Frequency))
				} else {
					//fileNotRuWords.WriteString(fmt.Sprintf("%s;%s;%s;%d\n", k, v.Lemma, v.POS, v.Frequency))
				}
			} else {
				if isRussian(resWord[0][0].Lemma) {
					res1[k] = resWord[0][0].Lemma
					//fileNoCompositeWords.WriteString(fmt.Sprintf("%s;%s;%s;%d\n", k, v.Lemma, v.POS, v.Frequency))
				} else {
					//fileNotRuWords.WriteString(fmt.Sprintf("%s;%s;%s;%d\n", k, v.Lemma, v.POS, v.Frequency))
				}
			}
		} else if len(resWord) == 0 {
			//fileZeroWords.WriteString(fmt.Sprintf("%s;%s;%s;%d\n", k, v.Lemma, v.POS, v.Frequency))
			res1[k] = k
		} else if len(resWord) > 1 {
			counterComposite++
			//fileCompositeWords.WriteString(fmt.Sprintf("%s;%s;%s;%d\n", k, v.Lemma, v.POS, v.Frequency))
			for _, lemma := range resWord {
				if len(lemma) == 2 {
					someWords = append(someWords, lemma[1].Lemma)
				} else {
					someWords = append(someWords, lemma[0].Lemma)
				}
			}
			//res1[k] = strings.Join(someWords, " ")
			//fmt.Println(res1[k])
			someWords = nil
		}
	}
	res2 := make(map[string]string)
	counterComposite2 := 0
	for k, v := range res1 {
		resWord = tokenizer.Normalize(v)
		if len(resWord) == 1 {
			if len(resWord[0]) == 2 {
				res2[k] = resWord[0][1].Lemma
			} else {
				res2[k] = resWord[0][0].Lemma
			}
		} else if len(resWord) == 0 {
			res2[k] = k
		} else if len(resWord) > 1 {
			counterComposite2++
			for _, lemma := range resWord {
				if len(lemma) == 2 {
					someWords = append(someWords, lemma[1].Lemma)
				} else {
					someWords = append(someWords, lemma[0].Lemma)
				}
			}
			res2[k] = strings.Join(someWords, " ")
			fmt.Println(k, res2[k])
			someWords = nil
		}
	}

	
	fmt.Println(counterComposite)
	fmt.Println(counterComposite2)
	f, err := os.OpenFile("diff1.csv", 0777, fs.FileMode(os.O_CREATE))
	defer f.Close()
	f.WriteString("word|lemma|lemma2\n")
	for k, v := range res1 {
		if v != res2[k] && v != "9" {
			f.WriteString(fmt.Sprintf("%s|%s|%s\n", k, v, res2[k]))
		}
	}
}


func TestNormalizer_Normalize_2(t *testing.T) {
	tokenizer := NewNormalizer()
	err := tokenizer.LoadDictionariesLocal("../data/words.csv_latest1.gz", "../spellcheck/spellcheck12.01.csv")
	if err != nil {
		log.Fatal(err)
	}

	var hitQueries []string
	fileHits, err := os.Open("../data/hits1000.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer fileHits.Close()

    scanner := bufio.NewScanner(fileHits)
    // optionally, resize scanner's capacity for lines over 64K, see next example
    i := 0
	for scanner.Scan() {
        line := scanner.Text()
		if line != "" {
			columns := strings.Split(line, " | ")
			if strings.Count(columns[0], " ") == len(columns[0]) {
				hitQueries[i-1] = hitQueries[i-1]+" "+strings.TrimLeft(strings.TrimRight(columns[1]," "), " ")
			} else {
				hitQueries = append(hitQueries, strings.TrimLeft(strings.TrimRight(columns[1]," "), " "))
				i++
			} 
		}
    }
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

	normalizationTokenizer := make(map[string]string, 1000) 
	sb := strings.Builder{}
	for _, query := range hitQueries {
		lemmas := tokenizer.Normalize(query)
		for i, lemma := range lemmas {
			if i >= len(lemmas)-1 {
				break
			}
			if len(lemma) == 2 {
				sb.WriteString(lemma[1].Lemma)
				sb.WriteRune(' ')
			} else {
				sb.WriteString(lemma[0].Lemma)
				sb.WriteRune(' ')
			}
		}
		if len(lemmas) > 0 {
			if len(lemmas[len(lemmas)-1]) == 2 {
				sb.WriteString(lemmas[len(lemmas)-1][1].Lemma)
			} else {
				sb.WriteString(lemmas[len(lemmas)-1][0].Lemma)
			}
		} 
		normalizationTokenizer[query] = sb.String()
		sb.Reset()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(hitQueries))
	wgPtr := &wg
	mutex := sync.Mutex{}
	mutexPtr := &mutex
	normalizationElastic := make(map[string]string, 1000) 
	mPtr := &normalizationElastic
	for _, query := range hitQueries {
		go ElasticNormalizationRequest(wgPtr, mPtr, mutexPtr, query)
	}
	wg.Wait()

	fmt.Println(normalizationTokenizer)
	fmt.Println(normalizationElastic["шапка женская с отворотом"])

	f, err := os.OpenFile("diff2.csv", 0777, fs.FileMode(os.O_CREATE))
	defer f.Close()
	f.WriteString("word|norm1|norm2\n")
	for k, v := range normalizationTokenizer {
		if v != normalizationElastic[k] {
			f.WriteString(fmt.Sprintf("%s|%s|%s\n", k, v, normalizationElastic[k]))
		}
	}
}

func ElasticNormalizationRequest(wg *sync.WaitGroup, m *map[string]string, mutex *sync.Mutex, word string) {
	request := ElasticNormRequest{
		Tokenizer : "standard",
  		Filter : []string{"lowercase",
						"protwords",
						"stopwords",
						"ru_morphology"},
  		Char_filter : []string{"ye_yo",
						"measurement_units",
						"size_units"},
  		Text : word,
	}
    json_data, err := json.Marshal(request)

    if err != nil {
        log.Fatal(err)
    }

    resp, err := http.Post("http://wbx-search-elastic-ru.dl.wb.ru:9200/nms-ru-000013/_analyze", "application/json",
        bytes.NewBuffer(json_data))

	var res ElasticNormResponse

	json.NewDecoder(resp.Body).Decode(&res)

	sb := strings.Builder{}
	for i := 0; i < len(res.Tokens)-1; i++ {
		sb.WriteString(res.Tokens[i].Token)
		sb.WriteRune(' ')
		
	}
	sb.WriteString(res.Tokens[len(res.Tokens)-1].Token)
	(*mutex).Lock()
	(*m)[word] = sb.String()
	(*mutex).Unlock()
	(*wg).Done()
} 