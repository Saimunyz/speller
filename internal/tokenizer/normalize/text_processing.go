package normalize

import (
	"strings"
	"unicode"
)

func simplifyPunctuation(runeText []rune) []rune {
	for i := 1; i < len(runeText)-1; i++ {
		if unicode.IsPunct(runeText[i]) {
			if runeText[i] == ',' || runeText[i] == '.' {
				// 2,5; 2.5
				if unicode.IsDigit(runeText[i-1]) && unicode.IsDigit(runeText[i+1]) {
					runeText[i] = '.'
					continue
				}
			} else if runeText[i] == '/' || runeText[i] == '&' {
				// a/b, a&b
				if unicode.IsLetter(runeText[i-1]) && unicode.IsLetter(runeText[i+1]) {
					continue
				}
			} else if runeText[i] == '*' {
				// 5*4 -> 5x4
				if unicode.IsDigit(runeText[i-1]) && unicode.IsDigit(runeText[i+1]) {
					runeText[i] = 'x'
					continue
				}
			} else if runeText[i] == '%' {
				// 100%
				if unicode.IsDigit(runeText[i-1]) {
					continue
				}
				// 100 %
				if i > 1 {
					if unicode.IsDigit(runeText[i-2]) && runeText[i-1] == ' ' {
						continue
					}
				}
			} else if runeText[i] == '-' {
				// -6 -> -6
				if unicode.IsDigit(runeText[i+1]) {
					continue
				}
			}
			runeText[i] = ' '
		} else if runeText[i] == '+' {
			// a+b -> a b ; a+6 -> a 6 ; 6+b -> 6 b
			if unicode.IsLetter(runeText[i+1]) || unicode.IsLetter(runeText[i-1]) {
				runeText[i] = ' '
				continue
			}
		}
	}
	if len(runeText) > 0 {
		if unicode.IsPunct(runeText[0]) && len(runeText) > 1 {
			if runeText[0] == '-' {
				if !unicode.IsDigit(runeText[1]) {
					runeText[0] = ' '
				}
			} else {
				runeText[0] = ' '
			}
		}
		if unicode.IsPunct(runeText[len(runeText)-1]) {
			if runeText[len(runeText)-1] != '%' {
				runeText[len(runeText)-1] = ' '
			}
		}
	}
	return runeText
}

func replaceSpaceSymbols(runeText []rune) []rune {
	for i := range runeText {
		if unicode.IsSpace(runeText[i]) {
			runeText[i] = ' '
		}
	}
	return runeText
}

func simplifyWhitespaces(runeText []rune) []rune {
	for i := 1; i < len(runeText); i++ {
		if unicode.IsSpace(runeText[i-1]) && unicode.IsSpace(runeText[i]) {
			runeText = append(runeText[:i], runeText[i+1:]...)
			i--
		}
	}
	return runeText
}

// 2м -> 2 м; 3тряпки -> 3 тряпки
func lostSpace(runeText []rune) []rune {
	result := strings.Builder{}
	result.WriteRune(runeText[0])
	for i := 1; i < len(runeText); i++ {
		if unicode.IsDigit(runeText[i-1]) && unicode.Is(unicode.Cyrillic, runeText[i]) {
			// 5х4 -> 5х4
			if runeText[i] == 'х' && i+1 < len(runeText) && unicode.IsDigit(runeText[i+1]) {
				result.WriteRune('х')
				continue
			}
			result.WriteRune(' ')
		}
		result.WriteRune(runeText[i])
	}
	return []rune(result.String())
}

func reductionToOneForm(runeText []rune) []rune {
	for i := range runeText {
		switch runeText[i] {
		case 'ё':
			runeText[i] = 'е'
		case 'é':
			runeText[i] = 'e'
		case 'ö':
			runeText[i] = 'o'
		}
	}
	for i := 1; i < len(runeText)-1; i++ {
		if unicode.IsDigit(runeText[i-1]) && unicode.IsDigit(runeText[i+1]) && runeText[i] == 'х' {
			runeText[i] = 'x'
		}
	}
	text := string(runeText)
	// гр/ -> г/
	text = strings.ReplaceAll(text, "гр/", "г/")

	// кв.м -> м2, куб.м -> м3
	return kvKubReduction([]rune(text))
}

// кв.м -> м2, куб.м -> м3
func kvKubReduction(chars []rune) []rune {
	var sb strings.Builder
	l := len(chars)
	sb.Grow(l)
	for i := 0; i < l; i++ {
		if chars[i] == 'к' {
			if i+1 < l {
				if chars[i+1] == 'в' {
					if i+2 < l && chars[i+2] == '.' {
						i = stringBuilderRussianRunesReduction(&sb, chars, i, l, "кв.", 3, "2")
					} else {
						sb.WriteString("кв")
						i += 1
					}
				} else if chars[i+1] == 'у' {
					if i+2 < l && chars[i+2] == 'б' {
						if i+3 < l && chars[i+3] == '.' {
							i = stringBuilderRussianRunesReduction(&sb, chars, i, l, "куб.", 4, "3")
						} else {
							sb.WriteString("куб")
							i += 2
						}
					} else {
						sb.WriteString("ку")
						i += 1
					}
				} else {
					sb.WriteRune('к')
				}
			} else {
				sb.WriteRune('к')
				break
			}
		} else {
			sb.WriteRune(chars[i])
		}
	}

	return []rune(sb.String())
}

//checks if current runes are russian letters, writes prefix to sb, add letters to sb, moves counter, returns counter
func stringBuilderRussianRunesReduction(sb *strings.Builder, chars []rune, i int, l int, prefix string,
	offset int, prefixReplacement string) int {
	if i+offset < l && chars[i+offset] <= 'я' && chars[i+offset] >= 'а' {
		sb.WriteRune(chars[i+offset])
	} else {
		sb.WriteString(prefix)
		return i + offset - 1
	}
	var j int
	for j = i + offset + 1; j < l && chars[j] <= 'я' && chars[j] >= 'а'; j++ {
		sb.WriteRune(chars[j])
	}
	sb.WriteString(prefixReplacement)
	return j - 1
}

func TextProcessing(text string) string {
	text = strings.ToLower(text)
	runeText := []rune(text)
	runeText = replaceSpaceSymbols(runeText)
	if len(text) > 0 {
		runeText = lostSpace(runeText)
		runeText = reductionToOneForm(runeText)
		runeText = simplifyPunctuation(runeText)
	}
	runeText = simplifyWhitespaces(runeText)
	return string(runeText)
}
