package strdist

// // DamerauLevenshtein distance is a string metric for measuring the edit
// // distance between two sequences:
// // https://en.wikipedia.org/wiki/Damerau%E3%80%93Levenshtein_distance
// //
// // This implementation has been designed using the observations of Steve
// // Hatchett:
// // http://blog.softwx.net/2015/01/optimizing-damerau-levenshtein_15.html
// //
// // Takes two strings and a maximum edit distance and returns the number of edits
// // to transform one string to another, or -1 if the distance is greater than the
// // maximum distance.

const (
	deletionWeight      = 0.6  //1.4//0.8
	replaceWeight       = 0.9  //1.5//0.9
	transpositionWeight = 0.8  //1.2//0.8
	insertWeight        = 1.01 //1.01
)

// DamerauLevenshteinRunes is the same as DamerauLevenshtein but accepts runes
// instead of strings
func KeyDamerauLevenshteinRunes(r1, r2 []rune, maxDist int) float64 {
	return DamerauLevenshteinRunesBuffer2(r1, r2, maxDist, nil, nil)
}

func DamerauLevenshteinRunesBuffer2(r1, r2 []rune, maxDist int, x, y []float64) float64 {

	if compareRuneSlices(r1, r2) {
		return 0
	}

	r1, r2, r1Len, r2Len, toReturn := swapRunes(r1, r2, maxDist)
	if toReturn != nil {
		return float64(*toReturn)
	}
	// save1 := make([]rune, len(r1))
	// save2 := make([]rune, len(r2))
	// copy(save1, r1)
	// copy(save2, r2)

	if r2Len-r1Len > maxDist {
		return -1
	}
	r1Len, r2Len = ignoreSuffix(r1, r2, r1Len, r2Len)

	// Ignore prefix
	start := 0
	if r1[start] == r2[start] || r1Len == 0 {

		for start < r1Len && r1[start] == r2[start] {
			start++
		}
		r1Len -= start
		r2Len -= start

		if r1Len == 0 {
			if r2Len <= maxDist {
				return float64(r2Len)
				// return float64(r2Len) * 0.8
			}
			return -1
		}
	}

	r2 = r2[start : start+r2Len]
	lenDiff, maxDist, toReturn := getLenDiff(r1Len, r2Len, maxDist)
	if toReturn != nil {
		return float64(*toReturn)
	}
	tableRunes := [][]rune{
		[]rune("впо"), //а
		[]rune("ью"),  //б
		[]rune("ыа"),  //в
		[]rune("нш"),  //г
		[]rune("лж"),  //д
		[]rune("кн"),  //е
		[]rune("дэ"),  //ж
		[]rune("щх"),  //з
		[]rune("мт"),  //и
		[]rune("фц"),  //й
		[]rune("уе"),  //к
		[]rune("од"),  //л
		[]rune("си"),  //м
		[]rune("ег"),  //н
		[]rune("рла"), //о
		[]rune("ар"),  //п
		[]rune("по"),  //р
		[]rune("чм"),  //с
		[]rune("иь"),  //т
		[]rune("цк"),  //у
		[]rune("йы"),  //ф
		[]rune("зъ"),  //х
		[]rune("йу"),  //ц
		[]rune("яс"),  //ч
		[]rune("гщ"),  //ш
		[]rune("шз"),  //щ
		[]rune("хэ"),  //ъ
		[]rune("фв"),  //ы
		[]rune("тб"),  //ь
		[]rune("жх"),  //э
		[]rune("б."),  //ю
		[]rune("фч"),  //я
	}
	x = getCharCosts(r2Len, maxDist, x)
	if y == nil {
		y = make([]float64, r2Len)
	}

	jStartOffset := maxDist - lenDiff
	haveMax := maxDist < r2Len // change to <
	jStart := 0
	jEnd := maxDist
	s1Char := r1[0]
	current := 0.0

	for i := 0; i < r1Len; i++ {
		prevS1Char := s1Char
		s1Char = r1[start+i]
		s2Char := r2[0]
		left := float64(i)
		current = left + 1
		nextTransCost := 0.0

		if i > jStartOffset {
			jStart++
		}

		if jEnd < r2Len {
			jEnd++
		}
		// var flag bool
		for j := jStart; j < jEnd; j++ {
			above := current
			thisTransCost := nextTransCost
			nextTransCost = y[j]
			current = left
			y[j] = current
			left = x[j]
			prevS2Char := s2Char //
			s2Char = r2[j]
			if s1Char != s2Char {
				if left < current { //insert
					current = left + insertWeight
				} else if above < current { //delete
					current = above + deletionWeight
				} else {
					// subst := getWeight(s1Char, s2Char, tableRunes);
					current += getWeight(s1Char, s2Char, tableRunes)
					// checkInd := s1Char - 'а'
					// if checkInd >= 0 && checkInd <= 31 {
					// 	// if variants := tableRunes[s1Char-'а']; variants[0] == s2Char || variants[1] == s2Char {
					// 	if tableRunes[checkInd][0] == s2Char || tableRunes[checkInd][1] == s2Char {
					// 		current += 0.2 //closeChange 0.2
					// 	} else {
					// 		// fmt.Println(string(save1), string(save2))
					// 		// time.Sleep(time.Second)
					// 		current += 1 //notCloseChange 0.6
					// 		// if flag {
					// 		// 	// fmt.Println(string(save1), string(save2))
					// 		// 	// time.Sleep(time.Second)
					// 		// 	return -1
					// 		// }
					// 		// flag = true
					// 	}
					// } else {
					// 	current += 1
					// }
					// current +=  1

				}

				// current = min(left * insertWeight, above * deletionWeight, current + getWeight(s1Char, s2Char, tableRunes))
				//current++
				if i != 0 && j != 0 && s1Char == prevS2Char && prevS1Char ==
					s2Char {
					thisTransCost++
					if thisTransCost < current {
						current = thisTransCost
					}
				}
			}
			x[j] = current
		}

		if haveMax && x[i+lenDiff] > float64(maxDist) { //float64(maxDist) { //-0.5
			return -1
		}
	}
	// fmt.Print(current, " ")
	return current
}

func getWeight(s1Char, s2Char rune, tableRunes [][]rune) float64 {
	checkInd := s1Char - 'а'
	if checkInd >= 0 && checkInd <= 31 {
		if len(tableRunes[checkInd]) == 3 {
			if tableRunes[checkInd][0] == s2Char || tableRunes[checkInd][1] == s2Char || tableRunes[checkInd][2] == s2Char {
				return 0.4 //closeChange
			} else {
				return 1 //notCloseChange
			}
		} else {
			if tableRunes[checkInd][0] == s2Char || tableRunes[checkInd][1] == s2Char {
				return 0.4 //closeChange
			} else {
				return 1 //notCloseChange
			}
		}

	}

	return 1
}
