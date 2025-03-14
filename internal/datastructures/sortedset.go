package datastructures

import (
	"container/heap"
	"fmt"
	"sort"
	"sync"
)

type SortedSet struct {
	mu      sync.RWMutex
	members map[string]float64
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		members: make(map[string]float64),
	}
}

func (s *SortedSet) Add(member string, score float64) {
	s.mu.Lockdefer s.mu.Unlock()
	s.members[member] = score
}

func (s *SortedSet) Remove(member string) {
	s.mu.Lockdefer s.mu.Unlock()
	delete(s.members, member)
}

func (s *SortedSet) GetScore(member string) (float64, bool) {
	s.mu.RLockdefer s.mu.RUnlock()
	score, exists := s.members[member]
	return score, exists
}

func (s *SortedSet) Range(min, max float64, offset, count int) []string {
	s.mu.RLockdefer s.mu.RUnlock()

	type scoredMember struct {
		score  float64
		member string
	}

	var members []scoredMember
	for member, score := range s.members {
		if score >= min && score <= max {
			members = append(members, scoredMember{score, member})
		}
	}

	// Sort by score
	sort.Slice(members, func(i, j int) bool {
		return members[i].score < members[j].score
	})

	// Apply offset and count
	start := offset
	if start > len(members) {
		start = len(members)
	}
	end := start + count
	if end > len(members) {
		end = len(members)
	}

	result := make([]string, 0, end-start)
	for _, member := range members[start:end] {
		result = append(result, member.member)
	}

	return result
}

func (s *SortedSet) ZCard() int {
	s.mu.RLockdefer s.mu.RUnlock()
	return len(s.members)
}

func (s *SortedSet) ZRank(member string) int {
	s.mu.RLockdefer s.mu.RUnlock()

	type scoredMember struct {
		score  float64
		member string
	}

	var members []scoredMember
	for m, score := range s.members {
		members = append(members, scoredMember{score, m})
	}

	// Sort by score
	sort.Slice(members, func(i, j int) bool {
		return members[i].score < members[j].score
	})

	// Find rank
	for i, m := range members {
		if m.member == member {
			return i
		}
	}

	return -1
}

func (s *SortedSet) ZRevRange(min, max float64, offset, count int) []string {
	s.mu.RLockdefer s.mu.RUnlock()

	type scoredMember struct {
		score  float64
		member string
	}

	var members []scoredMember
	for member, score := range s.members {
		if score >= min && score <= max {
			members = append(members, scoredMember{score, member})
		}
	}

	// Sort by score in descending order
	sort.Slice(members, func(i, j int) bool {
		return members[i].score > members[j].score
	})

	// Apply offset and count
	start := offset
	if start > len(members) {
		start = len(members)
	}
	end := start + count
	if end > len(members) {
		end = len(members)
	}

	result := make([]string, 0, end-start)
	for _, member := range members[start:end] {
		result = append(result, member.member)
	}

	return result
}

func (s *SortedSet) ZIntersect(sets []*SortedSet, weights []float64, aggregate string) *SortedSet {
	result := NewSortedSet()

	if len(sets) == 0 {
		return result
	}

	if len(weights) == 0 {
		weights = make([]float64, len(sets))
		for i := range weights {
			weights[i] = 1.0
		}
	}

	// Use a heap to efficiently get the next member with highest score
	type scoredMember struct {
		score  float64
		member string
	}

	// For simplicity, we'll use a map to track scores
	scores := make(map[string]float64)

	for i, set := range sets {
		members := set.Range(0, float64(^uint64(0)>>1), 0, set.ZCard())
		for _, member := range members {
			score, _ := set.GetScore(member)
			weightedScore := score * weights[i]

			if existingScore, exists := scores[member]; exists {
				switch aggregate {
				case "SUM":
					scores[member] = existingScore + weightedScore
				case "MIN":
					if weightedScore < existingScore {
						scores[member] = weightedScore
					}
				case "MAX":
					if weightedScore > existingScore {
						scores[member] = weightedScore
					}
				}
			} else {
				scores[member] = weightedScore
			}
		}
	}

	// Add all aggregated scores to the result set
	for member, score := range scores {
		result.Add(member, score)
	}

	return result
}