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
	re_YMD = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
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

func (s *InputParser) write() error {
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
		filename := fmt.Sprintf("Thai %s.txt", key)
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

func (s *InputParser) parse() error {
	s.lessons = make(map[string]*QuizletLesson)

	fh, err := os.Open("/mnt/chromeos/MyFiles/Downloads/Thai Vocabulary 2023.txt")
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
		if len(fields) == 0 {
			// Blank line
			continue
		} else if len(fields) == 1 {
			// YYYY-MM-DD line, hopefully
			token := fields[0]
			if re_YMD.MatchString(token) {
				// 7 bytes for YYYY-MM
				newKey := token[0:7]
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
			// We are expeting at least 2 spaces as the separator
			i := strings.Index(line, "  ")
			if i == -1 {
				panic(fmt.Sprintf("Expected a multi-space separator on line %d, saw: %s",
					lineno, line))
			}
			lhs := strings.TrimSpace(line[:i])
			rhs := strings.TrimSpace(line[i+2:])
			s.thisLesson.Add(lhs, rhs)
		}

	}
	err = scanner.Err()
	return err
}

func (s *Program) convert() error {

	var parser InputParser

	err := parser.parse()
	if err != nil {
		return err
	}

	fmt.Printf("Got %d lessons\n", len(parser.lessons))
	err = parser.write()
	if err != nil {
		return err
	}

	return nil
}
