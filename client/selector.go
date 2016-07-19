package client

import "time"
import "sync"

type DutyDecider interface {
	ChangedSince(time.Time) bool
	IsDutyBlock(int) bool
}

type selector struct {
	sync.Mutex
	decider          DutyDecider
	deciderTimestamp time.Time
	marks            []bool
	duties           []int
	n                int
	others           []int
	m                int
}

func newSelector(decider DutyDecider, count int) *selector {
	s := &selector{
		decider: decider,
		marks:   make([]bool, count),
	}
	s.remarkAll()
	return s
}

func (s *selector) selectDuty() int {
	s.Lock()
	defer s.Unlock()
	if s.decider.ChangedSince(s.deciderTimestamp) {
		s.remarkAll()
	}
	if len(s.duties) == 0 {
		return -1
	}
	if s.n >= len(s.duties) {
		s.cleanDuties()
		s.n = 0
	}
	if s.n >= len(s.duties) {
		return -1
	}
	res := s.duties[s.n]
	s.n++
	return res
}

func (s *selector) selectOther() int {
	s.Lock()
	defer s.Unlock()
	if s.decider.ChangedSince(s.deciderTimestamp) {
		s.remarkAll()
	}
	if len(s.others) == 0 {
		return -1
	}
	if s.m >= len(s.others) {
		s.cleanOthers()
		s.m = 0
	}
	if s.m >= len(s.others) {
		return -1
	}
	res := s.others[s.m]
	s.m++
	return res
}

func (s *selector) remarkAll() {
	duties := make([]int, 0, len(s.marks))
	others := make([]int, 0, len(s.marks))
	for i := 0; i < len(s.marks); i++ {
		if s.marks[i] {
			continue
		} else if s.decider.IsDutyBlock(i) {
			duties = append(duties, i)
		} else {
			others = append(others, i)
		}
	}
	s.n = 0
	s.m = 0
	s.deciderTimestamp = time.Now()
}

func (s *selector) cleanDuties() {
	duties := make([]int, 0, len(s.duties))
	for _, block := range s.duties {
		if !s.marks[block] {
			duties = append(duties, block)
		}
	}
	s.duties = duties
}

func (s *selector) cleanOthers() {
	others := make([]int, 0, len(s.others))
	for _, block := range s.others {
		if !s.marks[block] {
			others = append(others, block)
		}
	}
	s.others = others
}
