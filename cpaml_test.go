package cpaml

import (
	"fmt"
	"testing"
)

func initTests() *Cpaml {

	idx1 := Init(13)

	idx1.AddToIndex("fox1", "The quick brown fox jumps over the lazy dog then once again runs away and calls 1234567890")
	idx1.AddToIndex("fox-frog", "The quick brown zebra jump over the lazy frog  then once again runs away and calls 8674334434")
	idx1.AddToIndex("i-run", "I once again run away and call 8674334434")

	return idx1
}

func TestLookupSimilar(t *testing.T) {

	idx1 := initTests()

	var tests = []struct {
		expectedId         string
		expectedSimilarity uint
		sample             string
	}{
		{"fox1", 100, "The quick brown fox jumps over the lazy dog then once again runs away and calls 1234567890"},
		{"fox1", 60, "The quick brown fox jumps over the lazy dog then once again runs away and calls noone"},
		{"fox1", 15, "The slow yellow fox jumps over the nasty fox then once again flies far away and calls me"},
		{"fox-frog", 100, "The quick brown zebra jump over the lazy frog  then once again runs away and calls 8674334434"},
		{"fox-frog", 30, "The quick brown zebra jump over the lazy frog"},
		{"i-run", 100, "The quick brown fox jumps over the lazy dog. I once again run away and call 8674334434"},
		{"", 0, "Lorem ipsum dolor sit amet, consectetur adipiscing elit"},
		{"i-run", 30, "Lorem ipsum dolor sit amet, consectetur adipiscing elit 8674334434"},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("expected: %s with similarity %d", tt.expectedId, tt.expectedSimilarity)
		t.Run(testName, func(t *testing.T) {
			phraseId, similarity := idx1.LookupSimilar(tt.sample)
			if phraseId != tt.expectedId || similarity < tt.expectedSimilarity {
				t.Errorf("got id=[%s] expected=[%s]. Got similarity=[%d] expected=[%d]. Text=%s", phraseId, tt.expectedId, similarity, tt.expectedSimilarity, tt.sample)
			}
		})
	}
}
