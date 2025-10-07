package strings

import (
	"fmt"
	"math/rand"
	"strings"
)

func RandomStructName(rng *rand.Rand) string {
	length := rng.Intn(maxLength - minLength + 1)

	gen := newStructNameGenerator(rng)
	return gen.Rnd(length)
}

const (
	letters   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	minLength = 9
	maxLength = 14
)

var (
	reservedKeywords = map[string]bool{
		"break": true, "default": true, "func": true, "interface": true,
		"select": true, "case": true, "defer": true, "go": true, "map": true,
		"struct": true, "chan": true, "else": true, "goto": true, "package": true,
		"switch": true, "const": true, "fallthrough": true, "if": true, "range": true,
		"type": true, "continue": true, "for": true, "import": true, "return": true,
		"var": true,
	}
)

func newStructNameGenerator(rng *rand.Rand) *randomStructNameGenerator {
	return &randomStructNameGenerator{
		//rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		rng: rng,
	}
}

type randomStructNameGenerator struct {
	rng *rand.Rand
}

func (r *randomStructNameGenerator) Rnd(length int) string {
	b := make([]byte, length)
	b[0] = letters[rand.Intn(len(letters))]
	for i := 1; i < length; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}

	name := fmt.Sprintf("Struct_%s_%d", string(b), rand.Intn(100000))
	name = strings.ReplaceAll(name, "-", "_")

	if reservedKeywords[strings.ToLower(name)] {
		name += "_X"
	}

	return name
}
