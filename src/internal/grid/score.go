package grid

import "strings"

func (g Grid) Score(lists [][]string) ([]int, [][]int) {
	solution := g.Solve()
	solutionSet := map[string]struct{}{}

	for _, word := range solution {
		solutionSet[word] = struct{}{}
	}

	wordCounts := map[string]int{}

	for _, list := range lists {
		for _, word := range list {
			word = strings.ToUpper(word)
			if _, ok := solutionSet[word]; ok {
				wordCounts[word] += 1
			}
		}
	}

	scores := [][]int{}
	totals := []int{}
	for i, list := range lists {
		totals = append(totals, 0)
		scores = append(scores, nil)

		for _, word := range list {
			word = strings.ToUpper(word)
			score := 0
			if wordCounts[word] == 1 {
				if len(word) <= 4 {
					score = 1
				} else if len(word) == 5 {
					score = 2
				} else if len(word) == 6 {
					score = 3
				} else if len(word) == 7 {
					score = 5
				} else if len(word) == 8 {
					score = 11
				} else {
					score = 2 * len(word)
				}
			}

			scores[i] = append(scores[i], score)
			totals[i] += score
		}
	}

	return totals, scores
}
