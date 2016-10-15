package grid

import "strings"

var scoreTable = [18]int{0, 0, 0, 1, 1, 2, 3, 5, 11, 18, 20, 22, 24, 26, 28, 30, 32, 34}

func (g Grid) Score(lists [][]string) ([]int, [][]int, []string, int, []int) {
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
				wordCounts[word]++
			}
		}
	}

	masterScore := make([]int, len(solution))
	masterTotal := 0
	for i, word := range solution {
		if wordCounts[word] > 0 {
			masterScore[i] = 0
		} else {
			masterScore[i] = scoreTable[len(word)]
		}
		masterTotal += masterScore[i]
	}

	scores := make([][]int, len(lists))
	totals := make([]int, len(lists))
	for i, list := range lists {
		totals[i] = 0
		scores[i] = make([]int, len(list))

		for j, word := range list {
			word = strings.ToUpper(word)
			if wordCounts[word] == 1 {
				scores[i][j] = scoreTable[len(word)]
			} else if _, ok := solutionSet[word]; !ok {
				scores[i][j] = -1
			} else {
				scores[i][j] = 0
			}
			totals[i] += scores[i][j]
		}
	}

	return totals, scores, solution, masterTotal, masterScore
}
