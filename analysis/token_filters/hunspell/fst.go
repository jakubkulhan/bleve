package hunspell

import (
	"fmt"
	"reflect"
)

const (
	initialCap = 1024
)

// see https://github.com/ikawaha/mast
// http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
type FST struct {
	prev         string
	buf          []*fstState
	dictionary   map[uint][]*fstState
	initialState *fstState
	states       []*fstState
	finalStates  []*fstState
}

func NewFST() (fst *FST) {
	fst = &FST{
		prev: "",
		buf: make([]*fstState, 1),
		dictionary: make(map[uint][]*fstState),
		states: make([]*fstState, 0, initialCap),
		finalStates: make([]*fstState, 0, initialCap),
	}

	fst.buf[0] = &fstState{}
	fst.buf[0].clear()

	return
}

type fstArc struct {
	state *fstState
	out   FSTOutput
}

type fstState struct {
	id          int
	transitions map[byte]*fstArc
	final       bool
}

func (a *fstState) equal(b *fstState) bool {
	if a == nil || b == nil {
		return false
	}

	if a.id != b.id || a.final != b.final || len(a.transitions) != len(a.transitions) {
		return false
	}

	for ch, aArc := range a.transitions {
		bArc, ok := b.transitions[ch]

		if !ok {
			return false
		}

		// TODO: optimize
		if aArc.state != bArc.state || !reflect.DeepEqual(aArc.out, bArc.out) {
			return false
		}
	}

	return true
}

func (s *fstState) hash() (h uint) {
	h = 0

	for b, v := range s.transitions {
		h += (uint(b) + uint(v.state.id)) * 1001
	}

	//	for _, v := range s.out {
	//		for _, vv := range v {
	//			h += uint(vv) * 117709
	//		}
	//	}

	return
}

func (s *fstState) clear() {
	s.id = 0
	s.transitions = make(map[byte]*fstArc)
	s.final = false
}

func (s *fstState) transitionTo(b byte, next *fstState) {
	arc, ok := s.transitions[b]
	if ok {
		arc.state = next
	} else {
		s.transitions[b] = &fstArc{state: next}
	}
}

func (fst *FST) Add(in string, out FSTOutput) error {
	if fst.prev > in {
		return fmt.Errorf("inputs added out of order, prev=%q, in=%q", fst.prev, in)
	} else if (fst.prev == in) {
		return fmt.Errorf("multiple inputs for %q", in)
	}

	if len(in) + 1 > len(fst.buf) {
		prevBufLen := len(fst.buf)
		newBuf := make([]*fstState, len(in) + 1)
		copy(newBuf, fst.buf)
		fst.buf = newBuf

		for i, l := prevBufLen, len(fst.buf); i < l; i++ {
			fst.buf[i] = &fstState{}
			fst.buf[i].clear()
		}
	}

	prefixLenPlus1 := commonPrefixLen(fst.prev, in) + 1

	for i := len(fst.prev); i >= prefixLenPlus1; i-- {
		state := fst.findMinimized(i)
		fst.buf[i - 1].transitionTo(fst.prev[i - 1], state)
	}

	for i, l := prefixLenPlus1, len(in); i <= l; i++ {
		fst.buf[i].clear()
		fst.buf[i - 1].transitionTo(in[i - 1], fst.buf[i])
	}

	if in != fst.prev {
		fst.buf[len(in)].final = true
	}

	for j := 1; j < prefixLenPlus1; j++ {
		arc := fst.buf[j - 1].transitions[in[j - 1]]

		if arc.out != nil {
			commonPrefix := arc.out.Common(out)
			wordSuffix := arc.out.Subtract(commonPrefix)
			arc.out = commonPrefix
			for _, nextArc := range fst.buf[j].transitions {
				if nextArc.out == nil {
					nextArc.out = wordSuffix
				} else {
					nextArc.out = nextArc.out.Prepend(wordSuffix)
				}
			}

			out = out.Subtract(commonPrefix)
		}
	}

	arc := fst.buf[prefixLenPlus1 - 1].transitions[in[prefixLenPlus1 - 1]]
	arc.out = out

	fst.prev = in

	return nil
}

func (fst *FST) Finish() error {
	for i := len(fst.prev); i > 0; i-- {
		state := fst.findMinimized(i)
		fst.buf[i - 1].transitionTo(fst.prev[i - 1], state)
	}

	fst.initialState = fst.buf[0]
	fst.addState(fst.buf[0], fst.buf[0].hash())

	return nil
}

func (fst *FST) Run(in string, acc FSTOutput) (out FSTOutput, ok bool) {
	var arc *fstArc
	state := fst.initialState

	out = acc

	for i, l := 0, len(in); i < l; i++ {
		arc, ok = state.transitions[in[i]]

		if !ok {
			return
		}

		if arc.out != nil {
			out = out.Append(arc.out)
		}

		state = arc.state
	}

	ok = state.final

	return
}

func (fst *FST) findMinimized(i int) (state *fstState) {
	h := fst.buf[i].hash()

	if possibleStates, ok := fst.dictionary[h]; ok {
		for _, possibleState := range possibleStates {
			if possibleState.equal(fst.buf[i]) {
				state = possibleState
				return
			}
		}
	}

	state = &fstState{}
	*state = *fst.buf[i]
	fst.buf[i].clear()

	fst.addState(state, h)

	return
}

func (fst *FST) addState(state *fstState, h uint) {
	state.id = len(fst.states)

	fst.states = append(fst.states, state)

	if state.final {
		fst.finalStates = append(fst.finalStates, state)
	}

	fst.dictionary[h] = append(fst.dictionary[h], state)
}

func commonPrefixLen(a, b string) (i int) {
	for i = 0; i < len(a) && i < len(b) && a[i] == b[i]; i++ {}
	return
}
