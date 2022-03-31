package spellcheck

import "testing"

func TestRevertLayout(t *testing.T) {
	input := "ыфьыгтп cvfhnajy"
	want := "samsung смартфон"
	if want != RevertLayout(input) {
		t.Fatalf("%s != %s", RevertLayout(input), want)
	}
}
