package grid

import (
	"math/rand"
	"time"
)

type Grid [4][4]string

var r *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))

func Generate(seedOutput *int64) Grid {
	seed := r.Int63()
	grid := GenerateFromSeed(seed)
	if seedOutput != nil {
		*seedOutput = seed
	}
	return grid
}

func GenerateFromSeed(seed int64) Grid {
	rand := rand.New(rand.NewSource(seed))
	grid := Grid{}
	i := 0
	for _, c := range rand.Perm(4 * 4) {
		grid[i%4][i/4] = cubes[c][rand.Intn(len(cubes[c]))]
		i++
	}
	return grid
}
