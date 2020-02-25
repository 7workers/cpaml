package cpaml

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type TextInIndex struct {
	MinCommon uint
	Id        string
	Text      string
	NofKmers  uint
}

type Cpaml struct {
	k                  int
	mtx                sync.RWMutex
	indexKmers         map[string]map[uint]uint
	allIndexedStrings  []TextInIndex
	id2Idx             map[string]uint
	percRepeatedDetect uint
	percMinKmersMatch  uint
}

func Init(kmerLength int) *Cpaml {
	c := Cpaml{k: kmerLength}
	c.percRepeatedDetect = 70
	c.percMinKmersMatch = 70
	c.allIndexedStrings = make([]TextInIndex, 0)
	c.allIndexedStrings = append(c.allIndexedStrings, TextInIndex{0, "", "", 0})
	c.id2Idx = make(map[string]uint)
	c.indexKmers = make(map[string]map[uint]uint)

	return &c
}

func (c *Cpaml) LookupSimilar(t string) (string, uint) {
	kmers, _ := c.kmerize(t, true)
	matches := make(map[uint]uint, 0)
	c.mtx.RLock()
	for t, nofKmersRepeats := range kmers {
		if nil == c.indexKmers[t] {
			continue
		}

		for idxText, nofKmerRepeatsInText := range c.indexKmers[t] {

			_, matchExists := matches[idxText]

			if !matchExists {
				matches[idxText] = 0
			}

			if nofKmersRepeats < nofKmerRepeatsInText {
				matches[idxText] += nofKmersRepeats
			} else {
				matches[idxText] += nofKmerRepeatsInText
			}
		}
	}
	c.mtx.RUnlock()

	var bestMatchId uint
	var bestMatchNofCommon uint

	for idxText, nofCommon := range matches {

		if nofCommon <= bestMatchNofCommon {
			continue
		}
		/*
			if c.allIndexedStrings[idxText].MinCommon > nofCommon {
				continue
			}
		*/
		bestMatchId = idxText
		bestMatchNofCommon = nofCommon
	}

	if bestMatchId == 0 {
		return "", 0
	}

	bestMatchedText := c.allIndexedStrings[int(bestMatchId)]

	bestMatchNofKmers := bestMatchedText.NofKmers

	prcText := 100 * bestMatchNofCommon / bestMatchNofKmers
	prcLookup := 100 * bestMatchNofCommon / uint(len(kmers))

	if prcText > prcLookup {
		return bestMatchedText.Id, prcText
	} else {
		return bestMatchedText.Id, prcLookup
	}
}

func (c *Cpaml) kmerize(t string, slide bool) (map[string]uint, uint) {
	t = regexp.MustCompile("[,.!?:;|…“/\\\\*\\s]").ReplaceAllString(t, "")
	t = strings.ToLower(t)
	t = regexp.MustCompile("[^\\p{L}\\p{N}]").ReplaceAllString(t, "")
	t = t + strings.Repeat("$", c.k-1)
	kmers := make(map[string]uint)
	lenNormalized := uint(len([]rune(t)))
	maxLen := int(int(lenNormalized) - c.k)
	windowSlideStep := c.k
	if slide {
		windowSlideStep = 1
	}
	for i := 0; i <= maxLen; i += windowSlideStep {
		v := string([]rune(t)[i:(i + c.k)])
		cnt, ok := kmers[v]
		if ok {
			kmers[v] = cnt + 1
		} else {
			kmers[v] = 1
		}
	}
	return kmers, lenNormalized
}

func (c *Cpaml) AddToSet(id string, t string) bool {
	_, found := c.id2Idx[id]
	if found {
		return false
	}
	isAdded, _ := c.AddToIndex(id, t)
	return isAdded
}

func (c *Cpaml) AddToIndex(id string, t string) (bool, bool) {

	c.mtx.Lock()
	idxText := uint(len(c.allIndexedStrings))
	kmers, lenNormalized := c.kmerize(t, false)
	minKmersRepeated := (lenNormalized / uint(c.k)) * c.percRepeatedDetect / 100

	if uint(len(kmers)) < minKmersRepeated {
		fmt.Println("repetitive string: " + t)
		fmt.Printf("kmers: %d len:%d \n", len(kmers), len(t))
		c.mtx.Unlock()
		return false, true
	}

	for kmer, nofKmers := range kmers {
		if nil == c.indexKmers[kmer] {
			c.indexKmers[kmer] = make(map[uint]uint, 0)
		}
		c.indexKmers[kmer][idxText] = nofKmers
	}

	var minKmers = uint(len(kmers) * int(c.percMinKmersMatch) / 100)

	t = ""

	c.allIndexedStrings = append(c.allIndexedStrings, TextInIndex{minKmers, id, t, uint(len(kmers))})
	c.id2Idx[id] = idxText
	c.mtx.Unlock()

	return true, false
}

func (c *Cpaml) RemoveFromIndex(idx uint, id string) {

	for kmer, mapIdxes := range c.indexKmers {

		_, found := mapIdxes[idx]

		if !found {
			continue
		}

		c.mtx.Lock()

		if len(mapIdxes) == 1 {
			delete(c.indexKmers, kmer)
			c.mtx.Unlock()
			continue
		}

		delete(mapIdxes, idx)
		c.mtx.Unlock()
	}

	c.mtx.Lock()
	delete(c.id2Idx, id)
	c.mtx.Unlock()
}

func (c *Cpaml) IsInIndex(id string) bool {
	_, found := c.id2Idx[id]
	return found
}

func (c *Cpaml) RemoveStale(isForRemove func(id string) bool) int {
	nofRemoved := 0
	for idInDb, idx := range c.id2Idx {
		if !isForRemove(idInDb) {
			continue
		}
		c.RemoveFromIndex(idx, idInDb)
		nofRemoved++
	}
	return nofRemoved
}

func (c *Cpaml) GetStats() {

}
