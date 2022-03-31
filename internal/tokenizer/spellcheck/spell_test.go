package spellcheck

import (
	"testing"
)

func TestSpellcheck(t *testing.T) {
	tests := [][2]string{{"{jkjlbkmybrb", "Холодильники"}, {"а52телеыон", "а 52 телефон"}, {"эаран", "аарон"},
		{"аажон", "аарон"}, {"яарон", "аарон"}, {"еерон", ""}, {"zfhjy", "аарон"}}
	s := New()

	// Use your file
	err := s.Init("spellcheck12.01.csv")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	var actual string
	passFlag := true
	for _, test := range tests {
		actual, _ = s.Lookup(test[0])
		if actual != test[1] {
			t.Logf("Test string: \"%s\" Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
	t.Log()
}
