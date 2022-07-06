package main

import (
	"testing"
)

func TestNDup(t *testing.T) {
	var output []byte
	got := string(ndup([]byte("test"), output, 2))
	want := string([]byte{116, 101, 115, 116, 116, 101, 115, 116})

	if got != want {
		t.Errorf("got %s want %s", string(got), string(want))
	}
}
