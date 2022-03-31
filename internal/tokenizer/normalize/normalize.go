package normalize

import (
	"sort"
	"strings"
	"unicode"
)

type LemmaWithMeta struct {
	Lemma string
	POS   string
	Prep  string
}

func (n *Normalizer) Normalize(text string) [][]LemmaWithMeta {
	return n.normalize(text)
}

func (n *Normalizer) NormalizeWithSort(text string) [][]LemmaWithMeta {
	result := n.normalize(text)
	sort.Slice(result, func(i, j int) bool {
		return result[i][0].Lemma < result[j][0].Lemma
	})
	return result
}

func (n *Normalizer) normalize(text string) [][]LemmaWithMeta {
	var result [][]LemmaWithMeta
	text = TextProcessing(text)
	words := strings.Split(text, " ")
	prep := ""
	for _, word := range words {
		if _, ok := prepositions[word]; ok {
			prep = word
			continue
		}
		if strings.Contains(word, "/") {
			if _, ok := dimensions[word]; ok {
				result = append(result, []LemmaWithMeta{{Lemma: word}})
				continue
			}
			parts := strings.Split(word, "/")
			for _, part := range parts {
				lemmas, ok := n.lemmatize(part, prep)
				if ok {
					prep = ""
				}
				if len(lemmas) > 0 {
					result = append(result, lemmas)
				}
			}
		} else {
			lemmas, ok := n.lemmatize(word, prep)
			if ok {
				prep = ""
			}
			if len(lemmas) > 0 {
				result = append(result, lemmas)
			}
		}
	}
	return result
}

func (n *Normalizer) lemmatize(word, prep string) ([]LemmaWithMeta, bool) {
	if !isRussian(word) {
		return []LemmaWithMeta{{Lemma: word}}, true
	}
	lemma, ok := n.getLemma(word)
	if !ok {
		// пробуем применить спеллкорректор
		if suggestion, found := n.spell.Lookup(word); found {
			lemma, ok = n.getLemma(suggestion)
			if !ok {
				return []LemmaWithMeta{{Lemma: word}}, true
			}
		} else {
			return []LemmaWithMeta{{Lemma: word}}, true
		}
	}
	words := []Word{*lemma}
	var result []LemmaWithMeta
	prepFound := false
	for i := range words {
		if words[i].Lemma != "" && words[i].Lemma != " " {
			if words[i].POS == "NOUN" && !prepFound {
				result = append(result, LemmaWithMeta{
					Lemma: words[i].Lemma,
					POS:   words[i].POS,
					Prep:  prep,
				})
				prepFound = true
			} else {
				result = append(result, LemmaWithMeta{
					Lemma: words[i].Lemma,
					POS:   words[i].POS,
				})
			}
		}
	}
	return result, prepFound
}

func (n *Normalizer) getLemma(word string) (*Word, bool) {
	if value, ok := specialLemmas[word]; ok {
		return &Word{Lemma: value}, true
	}
	if value, ok := acronyms[word]; ok {
		return &Word{Lemma: value, POS: "NOUN"}, true
	}
	if _, ok := stopWords[word]; ok {
		return &Word{}, true
	}
	if _, ok := prepositions[word]; ok {
		return &Word{Lemma: word}, true
	}
	if value, ok := n.words[word]; ok {
		return &value, true
	}
	return &Word{Lemma: word}, false
}

func isRussian(word string) bool {
	runeWord := []rune(word)
	for i := range runeWord {
		if !unicode.Is(unicode.Cyrillic, runeWord[i]) {
			return false
		}
	}
	return true
}

type LemmaWithBase struct {
	Base  string
	Lemma string
	POS   string
	Prep  string
}

func (n *Normalizer) NormalizeWithoutMeta(text string) [][]LemmaWithBase {
	var result [][]LemmaWithBase
	words := strings.Split(text, " ")
	prep := ""
	for _, word := range words {
		if _, ok := prepositions[word]; ok {
			result = append(result, []LemmaWithBase{{Base: word, Lemma: word}})
			continue
		}
		if strings.Contains(word, "/") {
			if _, ok := dimensions[word]; ok {
				result = append(result, []LemmaWithBase{{Base: word, Lemma: word}})
				continue
			}
			parts := strings.Split(word, "/")
			for _, part := range parts {
				lemmas, ok := n.lemmatizeWithoutDeletingStopWords(part, prep)
				if ok {
					prep = ""
				}
				if len(lemmas) > 0 {
					result = append(result, lemmas)
				}
			}
		} else {
			lemmas, ok := n.lemmatizeWithoutDeletingStopWords(word, prep)
			if ok {
				prep = ""
			}
			if len(lemmas) > 0 {
				result = append(result, lemmas)
			}
		}
	}
	return result
}

func (n *Normalizer) lemmatizeWithoutDeletingStopWords(word, prep string) ([]LemmaWithBase, bool) {
	if !isRussian(word) {
		return []LemmaWithBase{{Base: word, Lemma: word}}, true
	}
	lemma, ok := n.getLemmaWithoutDeletingStopWords(word)
	if !ok {
		// пробуем применить спеллкорректор
		if suggestion, found := n.spell.Lookup(word); found {
			lemma, ok = n.getLemmaWithoutDeletingStopWords(suggestion)
			if !ok {
				return []LemmaWithBase{{Base: word, Lemma: word}}, true
			}
		} else {
			return []LemmaWithBase{{Base:word, Lemma: word}}, true
		}
	}
	words := []Word{*lemma}
	var result []LemmaWithBase
	prepFound := false
	for i := range words {
		if words[i].Lemma != "" && words[i].Lemma != " " {
			if words[i].POS == "NOUN" && !prepFound {
				result = append(result, LemmaWithBase{
					Base:  word,
					Lemma: words[i].Lemma,
					POS:   words[i].POS,
					Prep:  prep,
				})
				prepFound = true
			} else {
				result = append(result, LemmaWithBase{
					Base:  word,
					Lemma: words[i].Lemma,
					POS:   words[i].POS,
				})
			}
		}
	}
	return result, prepFound
}

func (n *Normalizer) getLemmaWithoutDeletingStopWords(word string) (*Word, bool) {
	if value, ok := specialLemmas[word]; ok {
		return &Word{Lemma: value}, true
	}
	if value, ok := acronyms[word]; ok {
		return &Word{Lemma: value, POS: "NOUN"}, true
	}
	if _, ok := stopWords[word]; ok {
		return &Word{Lemma: word}, true
	}
	if _, ok := prepositions[word]; ok {
		return &Word{Lemma: word}, true
	}
	if value, ok := n.words[word]; ok {
		return &value, true
	}
	return &Word{Lemma: word}, false
}
