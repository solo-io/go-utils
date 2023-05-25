package helm

import (
	"strings"
)

func NewSpaces(s string) *Spaces {
	spaces := strings.Split(s, " ")
	return &Spaces{line: s, spaces: spaces}
}

type Spaces struct {
	line   string
	spaces []string
}

// Returns the number of spaces (' ') at the beginning of the line
func (s *Spaces) GetNumberOfSpacesAtBeginning() int {
	numberOfSpaces := 0
	for _, space := range s.spaces {
		if space == "" {
			numberOfSpaces++
		} else {
			break
		}
	}
	if s.ContainsOnlySpaces() {
		return numberOfSpaces - 1
	}
	return numberOfSpaces
}

// returns if the line starts as an YAML array
func (s *Spaces) BeginsWithArray() bool {
	for _, space := range s.spaces {
		if space == "" {
			continue
		} else if space == "-" {
			return true
		} else {
			return false
		}
	}
	return false
}

// returns if the line starts a new resource
func (s *Spaces) IsNewResource() bool {
	if len(s.spaces) == 0 {
		return false
	}
	return s.spaces[0] == "---"
}

// returns if the line starts with a comment
func (s *Spaces) StartsWithComment() bool {
	if len(s.spaces) == 0 {
		return false
	}
	// continue until the first space contains a comment, or contains anything that is not a comment
	for _, space := range s.spaces {
		if strings.HasPrefix(space, "#") {
			return true
		} else if space == "" {
			continue
		} else {
			return false
		}
	}
	return false
}

func (s *Spaces) IsEmptyLine() bool {
	if len(s.spaces) == 0 {
		return true
	}
	// only if there is a single entry and it is empty
	// also this could be a empty line with a single space in it,
	// the building of the helm chart will look for that, So it should not be a concern for us.
	// we will not be able to build a helm chart if there is a single/multiple spaces in a line only
	return s.spaces[0] == "" && len(s.spaces) == 1
}

// if the last section contians a bar dash (|-)
func (s *Spaces) HasSpecialBreak() bool {
	chars := s.spaces[len(s.spaces)-1]
	return chars == "|-" || chars == ">-" || chars == "|"
}

// returns true if the line only contains spaces
func (s *Spaces) ContainsOnlySpaces() bool {
	for _, space := range s.spaces {
		if space != "" {
			return false
		}
	}
	return true
}

func FindHelmChartWhiteSpaces(data string) [][]string {
	lines := strings.Split(string(data), "\n")
	// we want to count each line, if the number of spaces at the begining is equal to 0, +2, or -2 from the previous line
	// then we want to continue to the next line. Else we want to throw an error.
	previousNumOfSpaces := 0
	badWindows := [][]string{}
	specialBreak := false
	for currentIndex, line := range lines {
		s := NewSpaces(line)
		if s.StartsWithComment() {
			continue
		}
		// impact of this change, went from  20178 --> 10266 changes
		if s.IsEmptyLine() {
			continue
		}
		shouldContinue := false
		currentNumOfSpaces := s.GetNumberOfSpacesAtBeginning()
		beginsWithArray := s.BeginsWithArray()
		// next level is the next accpetable number of spaces
		nextLevel := previousNumOfSpaces + 2
		isCurrentLevel := previousNumOfSpaces == currentNumOfSpaces
		isNextLevel := currentNumOfSpaces == nextLevel
		isSmallerLevel := currentNumOfSpaces < previousNumOfSpaces

		// adding the barDash logic went from 10266 --> 546
		if isSmallerLevel && specialBreak {
			// we are now exiting the barDash
			specialBreak = false
		} else if specialBreak {
			// if we are in a barDash, just continue regardless,
			// until we exit the barDash
			continue
		}
		// // this means an empty line has occured, and it contains only spaces, so move on to the next line
		if (isCurrentLevel || isNextLevel) && s.ContainsOnlySpaces() {
			continue
		}
		// always set the previous number of spaces, regardless
		// this made the errors go from 546 --> 189
		if beginsWithArray && (isCurrentLevel || isSmallerLevel) {
			// add 2 to current level because the array can be on current level or next level
			// if on current level, the next line should be on the next level
			previousNumOfSpaces = currentNumOfSpaces + 2
		} else {
			previousNumOfSpaces = currentNumOfSpaces
		}

		if isCurrentLevel || isNextLevel || isSmallerLevel {
			shouldContinue = true
			// if the current is less than previous, we are moving out of an object or array
			// if it is an empty line we need to ignore it, nothing happens
		} else {
			windowLines := Window(lines, currentIndex, 6)
			badWindows = append(badWindows, windowLines)
			// for _, l := range windowLines {
			// 	fmt.Println(l)
			// }
			// panic(line)
		}
		// just record that there is a bar dash
		if s.HasSpecialBreak() {
			specialBreak = true
		}
		if shouldContinue {
			continue
		}
	}
	return badWindows
}

func Window(s []string, index int, windowSize int) []string {
	start := index - windowSize
	end := index + windowSize
	if len(s) < end {
		end = len(s)
	}
	if start < 0 {
		start = 0
	}
	return s[start:end]
}
