package cpaml

import (
	"regexp"
	"strings"
	"sync"
)

type TextInIndex struct {
	Id       string
	NofKmers uint
}

type Cpaml struct {
	k                 int
	mtx               sync.RWMutex
	indexKmers        map[string]map[uint]uint
	allIndexedStrings []TextInIndex
	id2Idx            map[string]uint
	prcRepeatedDetect uint
	rgxStrip          *regexp.Regexp
	tailString        string
}

type Stats struct {
	NofSamples      int
	NofKmersIndexed int
}

/**
Provide kmer length. Depend on your cases, 13 works well for English and user messages and comments.
Works not so good with short strings (100 signs and shorter)
*/
func Init(kmerLength int) *Cpaml {
	c := Cpaml{k: kmerLength}
	c.prcRepeatedDetect = 70
	c.allIndexedStrings = make([]TextInIndex, 0)
	c.allIndexedStrings = append(c.allIndexedStrings, TextInIndex{"", 0})
	c.id2Idx = make(map[string]uint)
	c.indexKmers = make(map[string]map[uint]uint)
	c.rgxStrip = regexp.MustCompile("[^\\p{L}\\p{N}]")
	c.tailString = strings.Repeat("$", c.k/2)
	return &c
}

/*
return sample ID as given for AddToIndex and similarity 0-100
*/
func (c *Cpaml) LookupSimilar(t string) (string, uint) {
	kmers, _ := c.kmerize(t, true)
	matches := make(map[uint]uint, 0)

	c.mtx.RLock()

	for kmerStr, nofKmersRepeats := range kmers {

		kmerRelated, found := c.indexKmers[kmerStr]

		if !found {
			continue
		}

		for idxText, nofKmerRepeatsInText := range kmerRelated {

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

	var bestMatchIdx uint
	var bestMatchPrc uint
	var bestMatchId string

	if len(matches) == 0 {
		return "", 0
	}

	for idxText, nofCommon := range matches {

		matchedText := &c.allIndexedStrings[int(idxText)]
		matchNofKmers := matchedText.NofKmers

		prcText := 100 * nofCommon / matchNofKmers
		prcLookup := 100 * nofCommon / uint(len(kmers))

		maxPrc := prcLookup

		if prcText > prcLookup {
			maxPrc = prcText
		}

		if maxPrc < bestMatchPrc {
			continue
		}

		if maxPrc == 100 {
			return matchedText.Id, 100
		}

		bestMatchIdx = idxText
		bestMatchPrc = maxPrc
		bestMatchId = matchedText.Id
	}

	if bestMatchIdx == 0 {
		return "", 0
	}

	return bestMatchId, bestMatchPrc
}

func (c *Cpaml) kmerize(t string, slide bool) (map[string]uint, uint) {
	t = c.rgxStrip.ReplaceAllString(t, "")
	t = strings.ToLower(t)
	t = t + c.tailString
	kmers := make(map[string]uint)
	lenNormalized := uint(len([]rune(t)))
	maxLen := int(int(lenNormalized) - c.k)
	windowSlideStep := c.k
	if slide {
		windowSlideStep = 1
	}
	runes := []rune(t)
	for i := 0; i <= maxLen; i += windowSlideStep {
		v := string(runes[i:(i + c.k)])
		cnt, ok := kmers[v]
		if ok {
			kmers[v] = cnt + 1
		} else {
			kmers[v] = 1
		}
	}
	return kmers, lenNormalized
}

/**
add sample to index if not added
retrun true if sample was added
return second true in case string cannot be added because has high kmer/length ration, mean repeated multiple times
*/
func (c *Cpaml) AddToSet(id string, t string) (bool, bool) {
	_, found := c.id2Idx[id]
	if found {
		return false, false
	}
	isAdded, isRepetitive := c.AddToIndex(id, t)
	return isAdded, isRepetitive
}

/**
Add sample to index
*/
func (c *Cpaml) AddToIndex(id string, t string) (bool, bool) {

	c.mtx.Lock()
	idxText := uint(len(c.allIndexedStrings))
	kmers, lenNormalized := c.kmerize(t, false)
	minKmersRepeated := (lenNormalized / uint(c.k)) * c.prcRepeatedDetect / 100

	if uint(len(kmers)) < minKmersRepeated {
		c.mtx.Unlock()
		return false, true
	}

	for kmer, nofKmers := range kmers {
		if nil == c.indexKmers[kmer] {
			c.indexKmers[kmer] = make(map[uint]uint, 0)
		}
		c.indexKmers[kmer][idxText] = nofKmers
	}

	c.allIndexedStrings = append(c.allIndexedStrings, TextInIndex{id, uint(len(kmers))})
	c.id2Idx[id] = idxText
	c.mtx.Unlock()

	return true, false
}

/**
remove from index. Recommended to use RemoveStale() instead
*/
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

/**
remove unused (inactive) samples from index. closure must return true for unused sample ID
*/
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

func (c *Cpaml) GetStats() Stats {
	return Stats{NofSamples: len(c.allIndexedStrings), NofKmersIndexed: len(c.indexKmers)}
}
