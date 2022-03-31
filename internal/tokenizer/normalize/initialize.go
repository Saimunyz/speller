package normalize

import (
	"bufio"
	"compress/gzip"
	"github.com/Saimunyz/speller/internal/tokenizer/spellcheck"
	"github.com/valyala/fasthttp"
	"os"
	"strconv"
	"strings"
)

type Normalizer struct {
	words map[string]Word
	spell *spellcheck.Spell
}

type Word struct {
	Word      string `json:"word"`
	POS       string `json:"pos"`
	Frequency int    `json:"frequency"`
	Lemma     string `json:"lemma"`
}

func NewNormalizer() *Normalizer {
	return &Normalizer{
		words: make(map[string]Word),
		spell: spellcheck.New(),
	}
}

const host = "https://images.wbstatic.net/wbx/tokenizator/"

func (n *Normalizer) LoadDictionariesLocal(pathDict string, pathSpellcheck string) error {
	err := n.spell.Init(pathSpellcheck)
	if err != nil {
		return err
	}
	file, err := os.Open(pathDict)
	if err != nil {
		return err
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gr.Close()
	scanner := bufio.NewReader(gr)
	_, _, err = scanner.ReadLine()
	if err != nil {
		return err
	}
	var (
		fr      int
		line    []byte
		lineErr error
	)
	for {
		line, _, lineErr = scanner.ReadLine()
		if lineErr != nil {
			break
		}
		fields := strings.Split(string(line), ";")
		if fields[3] == "" {
			fr = 0
		} else {
			fr, _ = strconv.Atoi(fields[3])
		}
		word := Word{
			Word:      fields[0],
			POS:       fields[2],
			Frequency: fr,
			Lemma:     fields[1],
		}
		n.words[word.Word] = word
	}
	return nil
}

func (n *Normalizer) LoadDictionariesRemote(dict, spell string) error {
	data, err := getData(spell)
	if err != nil {
		return err
	}
	err = n.spell.InitRemote(data)
	if err != nil {
		return err
	}
	data, err = getDataGziped(dict)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	var fr int
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ";")
		if fields[3] == "" {
			fr = 0
		} else {
			fr, _ = strconv.Atoi(fields[3])
		}
		word := Word{
			Word:      fields[0],
			POS:       fields[2],
			Frequency: fr,
			Lemma:     fields[1],
		}
		n.words[word.Word] = word
	}
	return nil
}

func getData(file string) ([]byte, error) {
	httpReq := fasthttp.AcquireRequest()
	httpResp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(httpReq)
		fasthttp.ReleaseResponse(httpResp)
	}()
	httpReq.SetRequestURI(host + file)
	httpReq.Header.SetMethod("GET")
	if err := fasthttp.Do(httpReq, httpResp); err != nil {
		return nil, err
	}
	return httpResp.Body(), nil
}

func getDataGziped(file string) ([]byte, error) {
	httpReq := fasthttp.AcquireRequest()
	httpResp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(httpReq)
		fasthttp.ReleaseResponse(httpResp)
	}()
	httpReq.SetRequestURI(host + file)
	httpReq.Header.SetMethod("GET")
	if err := fasthttp.Do(httpReq, httpResp); err != nil {
		return nil, err
	}
	return httpResp.BodyGunzip()
}
