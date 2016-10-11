package nickname

import (
	"io/ioutil"
	"math/rand"
	"path"
	"strings"

	"internal/log"
)

var adjectives []string
var animals []string

func init() {
	var err error

	if adjectives, err = load(path.Join("config", "adjectives.list")); err != nil {
		log.Fields{"error": err}.Panic("unable to load nickname list")
	}

	if animals, err = load(path.Join("config", "animals.list")); err != nil {
		log.Fields{"error": err}.Panic("unable to load nickname list")
	}
}

func load(path string) ([]string, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.Title(strings.Trim(string(blob), "\n")), "\n"), nil
}

func Generate() string {
	return adjectives[rand.Intn(len(adjectives))] + " " + animals[rand.Intn(len(animals))]
}
