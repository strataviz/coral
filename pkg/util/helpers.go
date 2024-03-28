package util

import (
	"crypto/md5"
	"fmt"
	"math/rand"
)

// TODO: Make this configurable later.
const (
	LabelPrefix = "image.stvz.io/"
)

func AppendFunc[S, T any](a []S, b []T, fn func(S, T) S) []S {
	var result []S
	for _, item := range a {
		for _, other := range b {
			result = append(result, fn(item, other))
		}
	}

	return result
}

func FilterFunc[S, T any](a []S, b []T, fn func(S, T) bool) []S {
	var result []S
	for _, i := range a {
		for _, j := range b {
			if fn(i, j) {
				result = append(result, i)
			}
		}
	}

	return result
}

func ImageLabelKey(hash string) string {
	return fmt.Sprintf("%s%s", LabelPrefix, hash)
}

func ImageHasher(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func RandString(n int) string {
	b := make([]rune, n)
	chars := []rune("abcdefghijklmnopqrstuvwxyz1234567890")

	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}

func SetIntersection(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}

	var result []string
	for _, item := range b {
		if _, ok := m[item]; ok {
			result = append(result, item)
		}
	}

	return result
}

func SetDifference(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range b {
		m[item] = true
	}

	var result []string
	for _, item := range a {
		if _, ok := m[item]; !ok {
			result = append(result, item)
		}
	}

	return result
}
