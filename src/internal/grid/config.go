package grid

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"internal/log"
	"internal/wordlist"
)

type cubeSet [4 * 4][6]string

var list wordlist.Wordlist
var cubes cubeSet

func init() {
	var err error
	if list, err = wordlist.FromFile(path.Join("config", "words.list")); err != nil {
		log.Fields{"error": err}.Panic("couldn't load wordlist")
	}

	var cubeData []byte
	if cubeData, err = ioutil.ReadFile(path.Join("config", "cubes.json")); err != nil {
		log.Fields{"error": err}.Panic("couldn't read cubes")
	}

	if err = json.Unmarshal(cubeData, &cubes); err != nil {
		log.Fields{"error": err}.Panic("couldn't parse cubes")
	}

}
