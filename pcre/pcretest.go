// Copyright (C) 2011 Florian Weimer <fw@deneb.enyo.de>
// Copyright (C) 2022 unix-world.org

// GO Lang

// (c) 2022 unix-world.org
// v.20220317.0255

package pcre

import (
	"log"
)

func TestCompile() {
	var check = func(p string, groups int) {
		re, err := Compile(p, 0)
		if err != nil {
			log.Fatal(p, err)
		}
		if g := re.Groups(); g != groups {
			log.Fatal(p, g)
		}
	}
	check("", 0)
	check("^", 0)
	check("^$", 0)
	check("()", 1)
	check("(())", 2)
	check("((?:))", 1)
	log.Println("TestCompile() OK ...")
}

func TestCompileFail() {
	var check = func(p, msg string) {
		_, err := Compile(p, 0)
		switch {
		case err == nil:
			log.Fatal(p)
		case err.Error() != msg:
			log.Fatal(p, "Message:", err.Error())
		}
	}
	check("(",       "( (1): missing )")
	check(`\`,       `\ (1): \ at end of pattern`)
	check(`abc\`,    `abc\ (4): \ at end of pattern`)
	check("abc\000", "abc\000 (3): NUL byte in pattern")
	check("a\000bc", "a\000bc (1): NUL byte in pattern")
	log.Println("TestCompileFail() OK ...")
}

func TestMatcher() {
	var m Matcher
	checkmatch1 := func(m *Matcher, pattern, subject string, args ...interface{}) {
		re := MustCompile(pattern, 0)
		var prefix string
		if m == nil {
			m = re.Matcher([]byte(subject), 0)
		} else {
			m.Reset(re, []byte(subject), 0)
		}
		prefix = "[]byte"
		if len(args) == 0 {
			if m.Matches {
				log.Fatal(prefix, pattern, subject, "!Matches")
			}
		} else {
			if !m.Matches {
				log.Fatal(prefix, pattern, subject, "Matches")
				return
			}
			if m.Groups != len(args)-1 {
				log.Fatal(prefix, pattern, subject, "Groups", m.Groups)
				return
			}
			for i, arg := range args {
				if s, ok := arg.(string); ok {
					if !m.Present(i) {
						log.Fatal(prefix, pattern, subject,
							"Present", i)

					}
					if g := string(m.Group(i)); g != s {
						log.Fatal(prefix, pattern, subject,
							"Group", i, g, "!=", s)
					}
				} else {
					if m.Present(i) {
						log.Fatal(prefix, pattern, subject,
							"!Present", i)
					}
				}
			}
		}
	}
	check := func(pattern, subject string, args ...interface{}) {
		checkmatch1(nil, pattern, subject, args...)
		checkmatch1(nil, pattern, subject, args...)
		checkmatch1(&m, pattern, subject, args...)
		checkmatch1(&m, pattern, subject, args...)
	}
	check(`^$`, "", "")
	check(`^abc$`, "abc", "abc")
	check(`^(X)*ab(c)$`, "abc", "abc", nil, "c")
	check(`^(X)*ab()c$`, "abc", "abc", nil, "")
	check(`^.*$`, "abc", "abc")
	check(`^.*$`, "a\000c", "a\000c")
	check(`^(.*)$`, "a\000c", "a\000c", "a\000c")
	log.Println("TestMatcher() OK ...")
}

func TestPartial() {
	re := MustCompile(`^abc`, 0)
	// Check we get a partial match when we should
	m1 := re.Matcher([]byte("ab"), PARTIAL_SOFT)
	if !m1.Matches {
		log.Fatal("Failed to find any matches")
	} else if !m1.Partial {
		log.Fatal("The match was not partial")
	}
	// Check we get an exact match when we should
	m2 := re.Matcher([]byte("abc"), PARTIAL_SOFT)
	if !m2.Matches {
		log.Fatal("Failed to find any matches")
	} else if m2.Partial {
		log.Fatal("The match was net partial")
	}
	log.Println("TestPartial() OK ...")
}

func TestCaseless() {
	m := MustCompile("abc", CASELESS).Matcher([]byte("Abc"), 0)
	if !m.Matches {
		log.Fatal("CASELESS")
	}
	m = MustCompile("abc", 0).Matcher([]byte("Abc"), 0)
	if m.Matches {
		log.Fatal("!CASELESS")
	}
	log.Println("TestCaseless() OK ...")
}

func TestNamed() {
	m := MustCompile("(?<L>a)(?<M>X)*bc(?<DIGITS>\\d*)", 0).Matcher([]byte("abc12"), 0)
	if !m.Matches {
		log.Fatal("Matches")
	}
	if !m.NamedPresent("L") {
		log.Fatal("NamedPresent(\"L\")")
	}
	if m.NamedPresent("M") {
		log.Fatal("NamedPresent(\"M\")")
	}
	if !m.NamedPresent("DIGITS") {
		log.Fatal("NamedPresent(\"DIGITS\")")
	}
	group, err := m.Named("DIGITS")
	if err != nil || "12" != string(group) {
		log.Fatal("NamedPresent(\"DIGITS\") == 12")
	}
	log.Println("TestNamed() OK ...")
}

func TestFindIndex() {
	re := MustCompile("bcd", 0)
	i := re.FindIndex([]byte("abcdef"), 0)
	if i[0] != 1 {
		log.Fatal("FindIndex start", i[0])
	}
	if i[1] != 4 {
		log.Fatal("FindIndex end", i[1])
	}
	log.Println("TestFindIndex() OK ...")
}

func TestReplaceAll() {
	re := MustCompile("foo", 0)
	// Don't change at ends.
	result := re.ReplaceAll([]byte("I like foods."), []byte("car"), 0)
	if string(result) != "I like cards." {
		log.Fatal("ReplaceAll", result)
	}
	// Change at ends.
	result = re.ReplaceAll([]byte("food fight fools foo"), []byte("car"), 0)
	if string(result) != "card fight carls car" {
		log.Fatal("ReplaceAll2", result)
	}
	log.Println("TestReplaceAll() OK ...")
}

func TestExtract() {
	re := MustCompile("b(c)(d)", 0)
	m := re.Matcher([]byte("abcdef"), 0)
	i := m.Extract()
	if string(i[0]) != "bcd" {
		log.Fatal("Full line unavailable: ", string(i[0]))
	}
	if string(i[1]) != "c" {
		log.Fatal("First match group not as expected: ", string(i[1]))
	}
	if string(i[2]) != "d" {
		log.Fatal("Second match group not as expected: ", string(i[2]))
	}
	log.Println("TestExtract() OK ...")
}

func TestExtractString() {
	tpl := `
[%%%IF:XA:==###123xYz###;(1)%%%]
A コアテスト·スイート
[%%%ELSE:XA(1)%%%]
B
[%%%/IF:XA(1)%%%]

[%%%IF:XA:==###123xYz###;(1)%%%]
C
[%%%ELSE:XA(1)%%%]
D
[%%%/IF:XA(1)%%%]
	`
	re := MustCompile(`(?sU)\[%%%IF\:([a-zA-Z0-9_\-\.]+)\:(@\=\=|@\!\=|@\<\=|@\<|@\>\=|@\>|\=\=|\!\=|\<\=|\<|\>\=|\>|\!%|%|\!\?|\?|\^~|\^\*|&~|&\*|\$~|\$\*)([^\[\]]*);((\([0-9]+\))?)%%%\](.*)?(\[%%%ELSE\:\1\4%%%\](.*)?)?\[%%%\/IF\:\1\4%%%\]`, 0)
	m := re.Matcher([]byte(tpl), 0)
	i := m.ExtractString()
	if i[0] != "[%%%IF:XA:==###123xYz###;(1)%%%]\nA コアテスト·スイート\n[%%%ELSE:XA(1)%%%]\nB\n[%%%/IF:XA(1)%%%]" {
		log.Fatal("Full line is wrong: ", "`" + i[0] + "`")
	}
	if i[1] != "XA" {
		log.Fatal("Group 1 is wrong: ", "`" + i[1] + "`")
	}
	if i[2] != "==" {
		log.Fatal("Group 2 is wrong: ", "`" + i[2] + "`")
	}
	if i[3] != "###123xYz###" {
		log.Fatal("Group 3 is wrong: ", "`" + i[3] + "`")
	}
	if i[4] != "(1)" {
		log.Fatal("Group 4 is wrong: ", "`" + i[4] + "`")
	}
	if i[5] != "(1)" {
		log.Fatal("Group 5 is wrong: ", "`" + i[5] + "`")
	}
	if i[6] != "\nA コアテスト·スイート\n" {
		log.Fatal("Group 6 is wrong: ", "`" + i[6] + "`")
	}
	if i[7] != "[%%%ELSE:XA(1)%%%]\nB\n" {
		log.Fatal("Group 7 is wrong: ", "`" + i[7] + "`")
	}
	if i[8] != "\nB\n" {
		log.Fatal("Group 8 is wrong: ", "`" + i[8] + "`")
	}
	log.Println("TestExtractString() OK ...")
}

// #END
