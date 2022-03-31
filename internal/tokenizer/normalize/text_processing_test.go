package normalize

import (
	"testing"
)

func TestTextProcessing(t *testing.T) {
	//text := "// желтый; кардиган 35 килограмм 20кг/см 12м"
	tests := [][]string{{"// желтый; кардиган 35*5 5х4 килограмм 20кг/см 12м",
		" желтый кардиган 35x5 5x4 килограмм 20 кг/см 12 м"}}
	var actual string
	passFlag := true
	for _, test := range tests {
		actual = TextProcessing(test[0])
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func TestSimplifyPunctuation(t *testing.T) {
	tests := [][]string{{"текст", "текст"}, {",текст,", " текст "}, {"-текст", " текст"}, {"-2,", "-2 "},
		{",текст , текст", " текст   текст"}, {"-2,5 2.5 ,", "-2.5 2.5  "}, {"100%", "100%"}, {"%", "%"},
		{"b/a , a&b ; 1&a b&1 1&1", "b/a   a&b   1 a b 1 1 1"}, {"текст 4*5,5x4 текст ", "текст 4x5.5x4 текст "},
		{" 100%;100 % ", " 100% 100 % "}, {" text+text 6+a a+6 6+6", " text text 6 a a 6 6+6"},
		{"ab-5-6-a-c", "ab-5-6 a c"}}
	var actual string
	passFlag := true
	for _, test := range tests {
		actual = string(simplifyPunctuation([]rune(test[0])))
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func TestReplaceSpaceSymbols(t *testing.T) {
	tests := [][]string{{"текст", "текст"}, {"текст\t \t", "текст   "},
		{"текст\t", "текст "}, {"\t\t\n\r\v\f \n ", "         "},
		{"\t  текст \n\r  \fтекст текст   \tтекст\tтекст  \n", "   текст      текст текст    текст текст   "}}
	var actual string
	passFlag := true
	for _, test := range tests {
		actual = string(replaceSpaceSymbols([]rune(test[0])))
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func TestSimplifyWhitespaces(t *testing.T) {
	tests := [][]string{{"текст", "текст"}, {"текст ", "текст "}, {"текст текст", "текст текст"}, {"  ", " "},
		{"текст   ", "текст "}, {"текст   текст", "текст текст"}, {"текст   ", "текст "},
		{"   текст      текст текст    текст текст   ", " текст текст текст текст текст "}}
	var actual string
	passFlag := true
	for _, test := range tests {
		actual = string(simplifyWhitespaces([]rune(test[0])))
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func TestLostSpace(t *testing.T) {
	tests := [][]string{{"2", "2"}, {"а", "а"}, {"2м", "2 м"}, {"2m", "2m"}, {"2 м", "2 м"},
		{"м2м 3тряпки 4", "м2 м 3 тряпки 4"}}
	var actual string
	passFlag := true
	for _, test := range tests {
		actual = string(lostSpace([]rune(test[0])))
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func TestReductionToOneForm(t *testing.T) {
	tests := [][]string{{"текст", "текст"}, {"тёкст", "текст"}, {"тéкст", "тeкст"}, {"тöкст", "тoкст"},
		{"2х5 2x5 2xА 2хA 2х2", "2x5 2x5 2xА 2хA 2x2"}, {"гр/л гp/л", "г/л гp/л"},
		{"тёксттéксттöкст2х5 2x5 2xА 2хA 2х2гр/л гp/л", "тексттeксттoкст2x5 2x5 2xА 2хA 2x2г/л гp/л"}}
	var actual string
	passFlag := true
	for _, test := range tests {
		actual = string(reductionToOneForm([]rune(test[0])))
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func TestReductionToOneFormKvKub(t *testing.T) {
	tests := [][]string{{"кв.см", "см2"}, {"куб.м", "м3"}, {"куб м", "куб м"}, {"кв.m", "кв.m"}, {"текст к", "текст к"},
		{"куб.n", "куб.n"}, {"текст кв.меm", "текст ме2m"}, {"текст куб.ме_тр текст ", "текст ме3_тр текст "},
		{"к текст к текст кв текст кв. текст ку текст куб текст куб. текст кв.метр текст куб.см текст к",
			"к текст к текст кв текст кв. текст ку текст куб текст куб. текст метр2 текст см3 текст к"}}

	var actual string
	passFlag := true
	for _, test := range tests {
		actual = string(kvKubReduction([]rune(test[0])))
		if actual != test[1] {
			t.Logf("Test string: \"%s\". Expected \"%s\", got \"%s\"", test[0], test[1], actual)
			passFlag = false
		}
	}
	if !passFlag {
		t.Fail()
	}
}

func BenchmarkBigTextReductionToOneForm(b *testing.B) {
	text := "кв.см куб.м куб. см гр/м ку куб.абв"
	for i := 0; i < 20; i++ {
		text += text
	}
	runeText := []rune(text)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reductionToOneForm(runeText)
	}
	b.Log()
}

func BenchmarkManyRequestsTextReductionToOneForm(b *testing.B) {
	text := "кв.см куб.м куб. см гр/м ку куб.абв"
	runeText := []rune(text)
	for j := 0; j < b.N; j++ {
		reductionToOneForm(runeText)
	}
	b.Log()
}
