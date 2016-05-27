package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"internal/grid"
	"internal/log"
	"internal/wordlist"

	"github.com/julienschmidt/httprouter"
)

var cubes grid.Cubes
var words wordlist.Wordlist

func init() {
	cubeData, err := ioutil.ReadFile(path.Join("config", "cubes.json"))
	if err != nil {
		log.Fields{"error": err}.Panic("couldn't read cubes")
	}

	if err = json.Unmarshal(cubeData, &cubes); err != nil {
		log.Fields{"error": err}.Panic("couldn't parse cubes")
	}

	words, err = wordlist.FromFile(path.Join("config", "words.list"))
	if err != nil {
		log.Fields{"error": err}.Panic("couldn't read wordlist")
	}
}

func gridHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var g grid.Grid
	var seed int64

	if seedParam := ps.ByName("seed"); seedParam != "" {
		var err error
		seed, err = strconv.ParseInt(seedParam, 10, 64)
		if err != nil {
			g = grid.Generate(cubes, &seed)
		} else {
			g = grid.GenerateFromSeed(cubes, seed)
		}
	} else {
		g = grid.Generate(cubes, &seed)
	}

	solution := g.Solve(words)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Grid #%d:\n%v\nFound %d words:\n%s", seed, g, len(solution), strings.Join(solution, "\n"))
}
