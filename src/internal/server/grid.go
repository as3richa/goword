package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"internal/grid"

	"github.com/julienschmidt/httprouter"
)

func gridHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var g grid.Grid
	var seed int64

	if seedParam := ps.ByName("seed"); seedParam != "" {
		var err error
		seed, err = strconv.ParseInt(seedParam, 10, 64)
		if err != nil {
			g = grid.Generate(&seed)
		} else {
			g = grid.GenerateFromSeed(seed)
		}
	} else {
		g = grid.Generate(&seed)
	}

	solution := g.Solve()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Grid #%d:\n%v\nFound %d words:\n%s", seed, g, len(solution), strings.Join(solution, "\n"))
}
