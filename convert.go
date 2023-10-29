package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	re_YMD = regexp.MustCompile(`\d{4}(/|-)\d{2}(/|-)\d{2}`)
)

type StringTuple struct {
	a string
	b string
}

type QuizletLesson struct {
	combinedSet Set[string]
	tuples      []*StringTuple
}

func (s *QuizletLesson) create() string {
	lines := make([]string, len(s.tuples))
	for i, tuple := range s.tuples {
		lines[i] = fmt.Sprintf("%s\t%s", tuple.a, tuple.b)
	}
	return strings.Join(lines, "\n")
}

func (s *QuizletLesson) Add(lhs, rhs string) {
	combined := lhs + "|" + rhs
	if s.combinedSet.Has(combined) {
		return
	}
	s.tuples = append(s.tuples, &StringTuple{
		a: lhs,
		b: rhs,
	})
}

type InputParser struct {

	// Key = "YYYY-MM"
	lessons    map[string]*QuizletLesson
	thisLesson *QuizletLesson
	thisKey    string

	parseError error
}

func (s *InputParser) write(prefix string) error {
	keys := make([]string, len(s.lessons))
	i := 0
	for key, _ := range s.lessons {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	for _, key := range keys {
		lesson := s.lessons[key]
		text := lesson.create()
		// TBD - check on disk to see if it changed
		filename := fmt.Sprintf("%s %s.txt", prefix, key)
		fmt.Printf("Writing %s\n", filename)
		fh, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer fh.Close()
		_, err = fh.WriteString(text)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *InputParser) parse(filename string) error {
	s.lessons = make(map[string]*QuizletLesson)

	fh, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fh.Close()
	scanner := bufio.NewScanner(fh)

	lineno := 0
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		// Downloaded from Google docs, the first line starts with the
		// Unicode BOM
		if lineno == 1 {
			// Check the first rune
			for _, r := range line {
				if r == '\uFEFF' {
					// We know it takes 3 bytes in UTF-8
					line = line[3:]
				}
				break
			}
		}

		fields := strings.Fields(line)

		// Break up a single "korean(english)" into 2 fields
		if len(fields) == 1 {
			lparen_i := strings.Index(line, "(")
			if lparen_i != -1 {
				fields = make([]string, 2)
				fields[0] = line[:lparen_i]
				fields[1] = line[lparen_i:]
			}
		}

		if len(fields) == 0 {
			// Blank line
			continue
		} else if len(fields) == 1 {
			// YYYY-MM-DD line, hopefully
			token := fields[0]
			if re_YMD.MatchString(token) {
				// 7 bytes for YYYY-MM
				// Ensure we have "-" not "/"
				newKey := token[0:4] + "-" + token[5:7]
				//				fmt.Printf("token=%s key=%s\n", token, newKey)
				if newKey == s.thisKey {
					continue
				} else {
					newLesson := &QuizletLesson{
						combinedSet: NewSet[string](),
					}
					s.thisLesson = newLesson
					s.lessons[newKey] = newLesson
					s.thisKey = newKey
					continue
				}
			} else {
				panic(fmt.Sprintf("Unexpected single field on line %d, saw: %s",
					lineno, line))
			}
		} else {
			// A new set of words
			if s.thisLesson == nil {
				panic(fmt.Sprintf("Encountered set of words before a date on line %d, saw: %s",
					lineno, line))
			}
			var found bool
			var lhs string
			var rhs string
			// Do we have a leaste 22 spaces as the separator
			if !found {
				i := strings.Index(line, "  ")
				if i != -1 {
					lhs = strings.TrimSpace(line[:i])
					rhs = strings.TrimSpace(line[i+2:])
					found = true
				}
			}
			// Maybe we have a colon separator
			if !found {
				i := strings.Index(line, ":")
				if i != -1 {
					lhs = strings.TrimSpace(line[:i])
					rhs = strings.TrimSpace(line[i+1:])
					found = true
				}
			}

			// Maybe the RHS is in parens
			if !found {
				paren_i := strings.Index(line, "(")
				paren_j := strings.LastIndex(line, ")")
				//fmt.Printf("line: %s ; paren_i=%d paren_j=%d\n", line, paren_i, paren_j)
				if paren_i != -1 && paren_j != -1 && paren_j > paren_i {
					lhs = strings.TrimSpace(line[:paren_i])
					rhs = strings.TrimSpace(line[paren_i+1 : paren_j])
					found = true
				}
			}

			if !found {
				panic(fmt.Sprintf("Expected a multi-space separator on line %d, saw: %s",
					lineno, line))
			}
			s.thisLesson.Add(lhs, rhs)
		}

	}
	err = scanner.Err()
	return err
}

func (s *Program) convert(filename string, prefix string) error {

	var parser InputParser

	err := parser.parse(filename)
	if err != nil {
		return err
	}

	fmt.Printf("Got %d lessons\n", len(parser.lessons))
	err = parser.write(prefix)
	if err != nil {
		return err
	}

	return nil
}
