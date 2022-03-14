package spellcorrect

import (
	"strings"
	"testing"
)

func TestSimpleTokenizerTokens(t *testing.T) {
	traindata := `Мама пришла домой поздно, но отец к тому времени ещё не вернулся`

	expected := []string{
		"мама", "пришла", "домой", "поздно", "но", "отец", "к", "тому",
		"времени", "ещё", "не", "вернулся",
	}

	reader := strings.NewReader(traindata)
	tokenizer := NewSimpleTokenizer()
	tokens, err := tokenizer.Tokens(reader)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if len(tokens) != len(expected) {
		t.Errorf("Not same len(expected)")
		return
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != tokens[i] {
			t.Errorf("token (%s) in position %d differ", expected[i], i)
			return
		}
	}
}
