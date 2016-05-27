package wordlist

import (
	"io/ioutil"
	"sort"
	"strings"
)

type Wordlist []string

func FromFile(path string) (Wordlist, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	words := strings.Split(strings.ToUpper(strings.Trim(string(blob), "\n")), "\n")
	sort.Strings(words)
	return words, nil
}
