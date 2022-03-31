package spellcheck

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

type Spell struct {
	dict map[string]string
}

func New() *Spell {
	s := new(Spell)
	s.dict = make(map[string]string)
	return s
}

func (s *Spell) Init(pathSpellcheck string) error {
	file, err := os.Open(pathSpellcheck)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	byteData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	lines := strings.Split(string(byteData), "\n")
	var lineData []string
	for i := 1; i < len(lines)-1; i++ {
		lineData = strings.Split(lines[i], ";")
		incorrectList := strings.Split(lineData[1], "|")
		for _, incorrectWord := range incorrectList {
			s.dict[incorrectWord] = lineData[0]
		}
	}
	return nil
}

func (s *Spell) InitRemote(blob []byte) error {
	r := bytes.NewReader(blob)
	reader := bufio.NewReader(r)
	byteData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	lines := strings.Split(string(byteData), "\n")
	var lineData []string
	for i := 1; i < len(lines)-1; i++ {
		lineData = strings.Split(lines[i], ";")
		incorrectList := strings.Split(lineData[1], "|")
		for _, incorrectWord := range incorrectList {
			s.dict[incorrectWord] = lineData[0]
		}
	}
	return nil
}

func (s *Spell) Lookup(input string) (string, bool) {
	if word, found := s.dict[input]; found {
		return word, true
	} else if word, found = s.dict[RevertLayout(input)]; found {
		return word, true
	}
	return "", false
}
