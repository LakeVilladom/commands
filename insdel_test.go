// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package commands

import (
	"reflect"
	"strings"
	"testing"

	"github.com/limetext/backend"
	"github.com/limetext/text"
)

func TestBasic(t *testing.T) {
	data := `Hello world
Test
Goodbye world
`
	ed := backend.GetEditor()
	w := ed.NewWindow()
	defer w.Close()

	v := w.NewFile()
	defer func() {
		v.SetScratch(true)
		v.Close()
	}()

	e := v.BeginEdit()
	v.Insert(e, 0, data)
	v.EndEdit(e)

	v.Sel().Clear()
	v.Sel().Add(text.Region{11, 11})
	v.Sel().Add(text.Region{16, 16})
	v.Sel().Add(text.Region{30, 30})
	ed.CommandHandler().RunTextCommand(v, "left_delete", nil)
	if v.Substr(text.Region{0, v.Size()}) != `Hello worl
Tes
Goodbye worl
` {
		t.Error(v.Substr(text.Region{0, v.Size()}))
	}
	ed.CommandHandler().RunTextCommand(v, "insert", backend.Args{"characters": "a"})
	if d := v.Substr(text.Region{0, v.Size()}); d != "Hello worla\nTesa\nGoodbye worla\n" {
		lines := strings.Split(v.Substr(text.Region{0, v.Size()}), "\n")
		for _, l := range lines {
			t.Errorf("%d: '%s'", len(l), l)
		}
	}

	v.Settings().Set("translate_tabs_to_spaces", true)
	ed.CommandHandler().RunTextCommand(v, "insert", backend.Args{"characters": "\t"})
	if v.Substr(text.Region{0, v.Size()}) != "Hello worla \nTesa    \nGoodbye worla   \n" {
		lines := strings.Split(v.Substr(text.Region{0, v.Size()}), "\n")
		for _, l := range lines {
			t.Errorf("%d: '%s'", len(l), l)
		}
	}
	ed.CommandHandler().RunTextCommand(v, "left_delete", nil)
	if d := v.Substr(text.Region{0, v.Size()}); d != "Hello worla\nTesa\nGoodbye worla\n" {
		lines := strings.Split(v.Substr(text.Region{0, v.Size()}), "\n")
		for _, l := range lines {
			t.Errorf("%d: '%s'", len(l), l)
		}
	}

	ed.CommandHandler().RunTextCommand(v, "left_delete", nil)
	if d := v.Substr(text.Region{0, v.Size()}); d != "Hello worl\nTes\nGoodbye worl\n" {
		lines := strings.Split(v.Substr(text.Region{0, v.Size()}), "\n")
		for _, l := range lines {
			t.Errorf("%d: '%s'", len(l), l)
		}
	}

	ed.CommandHandler().RunTextCommand(v, "insert", backend.Args{"characters": "\t"})
	if d := v.Substr(text.Region{0, v.Size()}); d != "Hello worl  \nTes \nGoodbye worl    \n" {
		lines := strings.Split(v.Substr(text.Region{0, v.Size()}), "\n")
		for _, l := range lines {
			t.Errorf("%d: '%s'", len(l), l)
		}
	}

	ed.CommandHandler().RunTextCommand(v, "left_delete", nil)
	if v.Substr(text.Region{0, v.Size()}) != "Hello worl\nTes\nGoodbye worl\n" {
		lines := strings.Split(v.Substr(text.Region{0, v.Size()}), "\n")
		for _, l := range lines {
			t.Errorf("%d: '%s'", len(l), l)
		}
	}

	edit := v.BeginEdit()
	v.Erase(edit, text.Region{0, v.Size()})
	v.Insert(edit, 0, "€þıœəßðĸʒ×ŋµåäö𝄞")
	v.EndEdit(edit)
	orig := "€þıœəßðĸʒ×ŋµåäö𝄞"
	if d := v.Substr(text.Region{0, v.Size()}); d != orig {
		t.Errorf("%s\n\t%v\n\t%v", d, []byte(d), []byte(orig))
	}

	v.Sel().Clear()
	v.Sel().Add(text.Region{3, 3})
	v.Sel().Add(text.Region{6, 6})
	v.Sel().Add(text.Region{9, 9})
	ed.CommandHandler().RunTextCommand(v, "left_delete", nil)
	exp := "€þœəðĸ×ŋµåäö𝄞"
	if d := v.Substr(text.Region{0, v.Size()}); d != exp {
		t.Errorf("%s\n\t%v\n\t%v", d, []byte(d), []byte(exp))
	}
}

type deleteTest struct {
	in, out []text.Region
	text    string
	ins     string
}

func runDeleteTest(command string, tests *[]deleteTest, t *testing.T) {
	ed := backend.GetEditor()
	w := ed.NewWindow()

	for i, test := range *tests {
		v := w.NewFile()
		defer func() {
			v.SetScratch(true)
			v.Close()
		}()

		e := v.BeginEdit()
		v.Insert(e, 0, test.ins)
		v.EndEdit(e)

		v.Sel().Clear()
		for _, r := range test.in {
			v.Sel().Add(r)
		}
		var s2 text.RegionSet
		for _, r := range test.out {
			s2.Add(r)
		}

		ed.CommandHandler().RunTextCommand(v, command, nil)
		if d := v.Substr(text.Region{0, v.Size()}); d != test.text {
			t.Errorf("Test %02d: Expected %s, but got %s", i, test.text, d)
		} else if !reflect.DeepEqual(*v.Sel(), s2) {
			t.Errorf("Test %02d: Expected %v, but have %v", i, s2, v.Sel())
		}
	}

}

func TestLeftDelete(t *testing.T) {
	tests := []deleteTest{
		{
			[]text.Region{{1, 1}, {2, 2}, {3, 3}, {4, 4}},
			[]text.Region{{0, 0}},
			"5678",
			"12345678",
		},
		{
			[]text.Region{{1, 1}, {3, 3}, {5, 5}, {7, 7}},
			[]text.Region{{0, 0}, {1, 1}, {2, 2}, {3, 3}},
			"2468",
			"12345678",
		},
		{
			[]text.Region{{1, 3}},
			[]text.Region{{1, 1}},
			"145678",
			"12345678",
		},
		{
			[]text.Region{{3, 1}},
			[]text.Region{{1, 1}},
			"145678",
			"12345678",
		},
		{
			[]text.Region{{100, 5}},
			[]text.Region{{93, 5}},
			"abc\nd",
			"abc\ndef\nghi\n",
		}, // Yes, this is indeed what ST3 does too.
	}

	runDeleteTest("left_delete", &tests, t)
}

func TestRightDelete(t *testing.T) {
	tests := []deleteTest{
		{
			[]text.Region{{0, 0}, {1, 1}, {2, 2}, {3, 3}},
			[]text.Region{{0, 0}},
			"5678",
			"12345678",
		},
		{
			[]text.Region{{1, 1}, {3, 3}, {5, 5}, {7, 7}},
			[]text.Region{{1, 1}, {2, 2}, {3, 3}, {4, 4}},
			"1357",
			"12345678",
		},
		{
			[]text.Region{{1, 3}},
			[]text.Region{{1, 1}},
			"145678",
			"12345678",
		},
		{
			[]text.Region{{3, 1}},
			[]text.Region{{1, 1}},
			"145678",
			"12345678",
		},
	}

	runDeleteTest("right_delete", &tests, t)
}

func TestInsert(t *testing.T) {
	ed := backend.GetEditor()
	ch := ed.CommandHandler()
	w := ed.NewWindow()
	defer w.Close()

	v := w.NewFile()
	defer func() {
		v.SetScratch(true)
		v.Close()
	}()

	e := v.BeginEdit()
	v.Insert(e, 0, "Hello World!\nTest123123\nAbrakadabra\n")
	v.EndEdit(e)

	type Test struct {
		in   []text.Region
		data string
		expd string
		expr []text.Region
	}

	tests := []Test{
		{
			[]text.Region{{1, 1}, {3, 3}, {6, 6}},
			"a",
			"Haelalo aWorld!\nTest123123\nAbrakadabra\n",
			[]text.Region{{2, 2}, {5, 5}, {9, 9}},
		},
		{
			[]text.Region{{1, 1}, {3, 3}, {6, 9}},
			"a",
			"Haelalo ald!\nTest123123\nAbrakadabra\n",
			[]text.Region{{2, 2}, {5, 5}, {9, 9}},
		},
		{
			[]text.Region{{1, 1}, {3, 3}, {6, 9}},
			"€þıœəßðĸʒ×ŋµåäö𝄞",
			"H€þıœəßðĸʒ×ŋµåäö𝄞el€þıœəßðĸʒ×ŋµåäö𝄞lo €þıœəßðĸʒ×ŋµåäö𝄞ld!\nTest123123\nAbrakadabra\n",
			[]text.Region{{17, 17}, {35, 35}, {54, 54}},
		},
	}

	for i, test := range tests {
		v.Sel().Clear()
		for _, r := range test.in {
			v.Sel().Add(r)
		}
		ed.CommandHandler().RunTextCommand(v, "insert", backend.Args{"characters": test.data})
		if d := v.Substr(text.Region{0, v.Size()}); d != test.expd {
			t.Errorf("Insert test %d failed: %s", i, d)
		}
		if sr := v.Sel().Regions(); !reflect.DeepEqual(sr, test.expr) {
			t.Errorf("Insert test %d failed: %v", i, sr)
		}
		ch.RunTextCommand(v, "undo", nil)
	}
}

func TestDeleteWord(t *testing.T) {
	tests := []struct {
		text    string
		sel     []text.Region
		forward bool
		expect  string
	}{
		{
			"word",
			[]text.Region{{4, 4}},
			false,
			"",
		},
		{
			"'(}[word",
			[]text.Region{{7, 7}, {4, 4}},
			false,
			"d",
		},
		{
			"testing forwar|d\ndelete word",
			[]text.Region{{0, 2}, {11, 11}, {16, 16}},
			true,
			"sting for|ddelete word",
		},
		{
			"simple 	test 	on outside",
			[]text.Region{{-1, -1}, {6, 6}, {13, 13}, {54, 33}, {31, 31}},
			true,
			"simpletest  outside",
		},
	}

	ed := backend.GetEditor()
	w := ed.NewWindow()
	defer w.Close()

	for i, test := range tests {
		v := w.NewFile()
		defer func() {
			v.SetScratch(true)
			v.Close()
		}()

		e := v.BeginEdit()
		v.Insert(e, 0, test.text)
		v.EndEdit(e)

		v.Sel().Clear()
		v.Sel().AddAll(test.sel)

		ed.CommandHandler().RunTextCommand(v, "delete_word", backend.Args{"forward": test.forward})
		if d := v.Substr(text.Region{0, v.Size()}); d != test.expect {
			t.Errorf("Test %d:\nExcepted: '%s' but got: '%s'", i, test.expect, d)
		}
	}
}
