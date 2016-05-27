package grid

import (
	"sort"

	"internal/wordlist"
)

type solveState struct {
	i, j, mask int
	query      string
}
type markTable map[solveState]bool

func (g Grid) Solve(list wordlist.Wordlist) []string {
	found := map[string]bool{}
	visited := markTable{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			g.recursiveSolve(visited, found, i, j, strike(0, i, j), list.NewSearch(g[i][j]))
		}
	}
	result := []string{}
	for key := range found {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func (g Grid) recursiveSolve(visited markTable, found map[string]bool, i, j, mask int, search wordlist.Search) {
	tuple := solveState{i, j, mask, search.Query}
	if visited[tuple] {
		return
	}
	visited[tuple] = true

	if len(search.Query) >= 3 && search.ExactMatch() {
		found[search.Query] = true
	}

	if !search.Empty() {
		for p := i - 1; p <= i+1; p++ {
			if !(0 <= p && p < 4) {
				continue
			}

			for q := j - 1; q <= j+1; q++ {
				if !(0 <= q && q < 4) {
					continue
				}

				if !struck(mask, p, q) {
					g.recursiveSolve(visited, found, p, q, strike(mask, p, q), search.Narrow(g[p][q]))
				}
			}
		}
	}
}

func strike(m, i, j int) int {
	return m | (1 << uint(i*4+j))
}

func struck(m, i, j int) bool {
	return (m & (1 << uint(i*4+j))) != 0
}
