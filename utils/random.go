package utils

import (
	"math/rand"
	"time"
)

// global random generator instance
var globalRand *rand.Rand

func init() {
	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomInt returns a random integer between min and max (inclusive)
func RandomInt(min, max int64) int64 {
	return min + globalRand.Int63n(max-min+1)
}

// RandomString returns a random string of length n
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[globalRand.Intn(len(letters))]
	}
	return string(result)
}
