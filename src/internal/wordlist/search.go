package wordlist

import (
	"sort"
	"strings"
)

type Search struct {
	list  Wordlist
	Query string
}

func (list Wordlist) NewSearch(initial string) Search {
	result := Search{
		list:  list,
		Query: "",
	}

	if len(initial) != 0 {
		return result.Narrow(initial)
	}

	return result
}

func (search Search) Narrow(delta string) Search {
	query := search.Query + strings.ToUpper(delta)

	left := earlyBinarySearch(search.list, query)
	list := search.list[left:]

	right := lateBinarySearch(list, query)
	list = list[:right]

	return Search{
		list:  list,
		Query: query,
	}
}

func (search Search) ExactMatch() bool {
	return (!search.Empty() && search.list[0] == search.Query)
}

func (search Search) Empty() bool {
	return (len(search.list) == 0)
}

func earlyBinarySearch(list Wordlist, query string) int {
	tentative := sort.SearchStrings(list, query)
	if tentative < len(list) && strings.HasPrefix(list[tentative], query) {
		return tentative
	}
	return len(list)
}

func lateBinarySearch(list Wordlist, query string) int {
	return len(list) - sort.Search(len(list), func(i int) bool {
		return strings.HasPrefix(list[len(list)-1-i], query)
	})
}
