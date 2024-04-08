// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

// #nosec

// TODO: Make this configurable later.

func ModifyFunc[S, T any](a []S, b []T, fn func(S, T) S) []S {
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

func FilterMapFunc[S, T comparable](a map[S]T, fn func(S, T) bool) map[S]T {
	result := make(map[S]T)

	for k, v := range a {
		if fn(k, v) {
			result[k] = v
		}
	}

	return result
}

func ListDiff(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		delete(m, item)
	}

	var result []string
	for k := range m {
		result = append(result, k)
	}

	return result
}
