package hunspell

import (
	"testing"
	"reflect"
	"os"
	"io"
	"fmt"
	"os/exec"
	"compress/gzip"
	"bufio"
	"math/rand"
)

func TestCommonPrefixLen(t *testing.T) {
	tests := []struct {
		a, b string
		l    int
	}{
		{"", "", 0},
		{"", "a", 0},
		{"a", "", 0},
		{"a", "a", 1},
		{"ab", "ac", 1},
		{"ac", "ab", 1},
		{"abc", "abd", 2},
		{"abcd", "abcd", 4},
	}

	for _, test := range tests {
		if got := commonPrefixLen(test.a, test.b); got != test.l {
			t.Errorf("commonPrefixLen(%q, %q) expected %v, got %v", test.a, test.b, test.l, got)
		}
	}
}

func TestEmptyFST(t *testing.T) {
	fst := NewFST()
	fst.Finish()

	fst.graph("_EmptyFST")

	if fst.initialState.id != 0 {
		t.Errorf("got initial state %v, expected 0", fst.initialState.id)
	}

	if len(fst.states) != 1 {
		t.Errorf("expected only 1 initial state, got %v", len(fst.states))
	}

	if len(fst.finalStates) != 0 {
		t.Errorf("expected no final state, got %v", len(fst.finalStates))
	}
}

func TestBasicFST(t *testing.T) {
	tests := []struct {
		in  string
		out Int32Output
	}{
		{"1", Int32Output([]int32{-1})},
		{"2", Int32Output([]int32{-2})},
		{"3", Int32Output([]int32{-3})},
		{"4", Int32Output([]int32{-4})},
		{"5", Int32Output([]int32{-5})},
		{"6", Int32Output([]int32{-6})},
		{"7", Int32Output([]int32{-7})},
		{"8", Int32Output([]int32{-8})},
		{"9", Int32Output([]int32{-9})},
		{"a", Int32Output([]int32{1})},
		{"abcd", Int32Output([]int32{2})},
		{"b", Int32Output([]int32{3})},
		{"bbcd", Int32Output([]int32{4})},
	}

	fst := NewFST()

	for _, test := range tests {
		fst.Add(test.in, test.out)
	}

	fst.Finish()

	fst.graph("_BasicFST")

	for _, test := range tests {
		if got, ok := fst.Run(test.in, Int32Output([]int32{})); !ok || !reflect.DeepEqual(got, test.out) {
			t.Errorf("wrong for %s; expected %t, %v; got %t, %v", test.in, true, test.out, ok, got)
		}
	}
}

func TestRepeatedInputFST(t *testing.T) {
	fst := NewFST()

	fst.Add("a", Int32Output([]int32{1}))
	fst.Add("a", Int32Output([]int32{2}))
	fst.Add("a", Int32Output([]int32{3}))

	fst.Finish()

	fst.graph("_RepeatedFST")

	got, ok := fst.Run("a", nil)

	if !ok {
		t.Errorf("expected ok")
	}

	expected := Int32Output([]int32{1, 2, 3})

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestMonthsFST(t *testing.T) {
	fst := NewFST()

	day28or29 := Int32Output([]int32{28, 29})
	day30 := Int32Output([]int32{30})
	day31 := Int32Output([]int32{31})

	months := []struct {
		month string
		days  Int32Output
	}{
		{"apr", day30},
		{"aug", day31},
		{"dec", day31},
		{"feb", day28or29},
		{"jan", day31},
		{"jul", day31},
		{"jun", day30},
		{"mar", day31},
		{"may", day31},
		{"nov", day30},
		{"oct", day31},
		{"sep", day30},
	}

	for _, month := range months {
		if err := fst.Add(month.month, month.days); err != nil {
			t.Error(err)
			return
		}
	}

	fst.Finish()

	fst.graph("_MonthsFST")

	for _, month := range months {
		if got, ok := fst.Run(month.month, Int32Output([]int32{})); !ok || !reflect.DeepEqual(got, month.days) {
			t.Errorf("wrong for %s, expected %v, got %v", month.month, month.days, got)
		}
	}
}

func Benchmark300kDictionary(b *testing.B) {
	dicGzip, err := os.Open("cs.dic.gz")
	if err != nil {
		b.Error(err)
		return
	}
	defer dicGzip.Close()

	dic, err := gzip.NewReader(dicGzip)
	if err != nil {
		b.Error(err)
		return
	}
	defer dic.Close()

	someWords := make([]string, 0, 1000)
	someInts := make([]int32, 0, cap(someWords))

	fst := NewFST()
	var x int32 = 0
	out := make(Int32Output, 1)

	for scanner := bufio.NewScanner(dic); scanner.Scan(); {
		word := scanner.Text()

		x++
		out[0] = x

		fst.Add(word, out)

		if len(someWords) < cap(someWords) && rand.Float32() > (1.0 / float32(cap(someWords) - len(someWords))) {
			someWords = append(someWords, word)
			someInts = append(someInts, x)
		}

		if x % 1000 == 0 {
			fmt.Printf("got %d\n", x)
		}
	}

	fst.Finish()

	b.ResetTimer()

	j := 0
	for i := 0; i < b.N; i++ {
		j++
		w := someWords[j % len(someWords)]
		x := someInts[j % len(someWords)]

		out, ok := fst.Run(w, out)

		if !ok {
			b.Errorf("word %s, expected ok, got not ok", w)
			continue
		}

		outInt32, ok := out.(Int32Output)

		if !ok {
			b.Errorf("expected Int32Output, got %T", outInt32)
			continue
		}

		if len(outInt32) != 1 || outInt32[0] != x{
			b.Errorf("word %s, expected %v, got %v", x, out)
			continue
		}
	}
}

func (fst *FST) graph(name string) error {
	fName := "../../../" + name + ".dot"
	f, err := os.OpenFile(fName, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	fst.dot(f)
	return exec.Command("dot", "-Tpng", fName, "-o", "../../../" + name + ".png").Run()
}

func (fst *FST) dot(w io.Writer) {
	fmt.Fprintln(w, "digraph G {")
	fmt.Fprintln(w, "\trankdir=LR;")
	fmt.Fprintln(w, "\tnode [shape=circle]")
	for _, s := range fst.finalStates {
		//		fmt.Fprintf(w, "\t%d [peripheries = 2 label=\"%d (%v)\"];\n", s.id, s.id, s.out)
		fmt.Fprintf(w, "\t%d [peripheries = 2];\n", s.id)
	}
	for _, from := range fst.states {
		for in, arc := range from.transitions {
			fmt.Fprintf(w, "\t%d -> %d [label=\"%c", from.id, arc.state.id, in)
			fmt.Fprintf(w, " %v", arc.out)
			fmt.Fprintln(w, "\"];")
		}
	}
	fmt.Fprintln(w, "}")
}
