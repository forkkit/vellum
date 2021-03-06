//  Copyright (c) 2017 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package regexp

import (
	"fmt"
	"testing"
)

func TestRegexp(t *testing.T) {
	tests := []struct {
		query    string
		seq      []byte
		isMatch  bool
		canMatch bool
	}{
		{
			query:    ``,
			seq:      []byte{},
			isMatch:  true,
			canMatch: true,
		},
		// test simple literal
		{
			query:    `a`,
			seq:      []byte{'a'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `a`,
			seq:      []byte{},
			isMatch:  false,
			canMatch: true,
		},
		{
			query:    `a`,
			seq:      []byte{'a', 'b'},
			isMatch:  false,
			canMatch: false,
		},
		// test actual pattern
		{
			query:    `wat.r`,
			seq:      []byte{'x'},
			isMatch:  false,
			canMatch: false,
		},
		{
			query:    `wat.r`,
			seq:      []byte{'w', 'a', 't'},
			isMatch:  false,
			canMatch: true,
		},
		{
			query:    `wat.r`,
			seq:      []byte{'w', 'a', 't', 'e'},
			isMatch:  false,
			canMatch: true,
		},
		{
			query:    `wat.r`,
			seq:      []byte{'w', 'a', 't', 'e', 'r'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `wat.r`,
			seq:      []byte{'w', 'a', 't', 'e', 'r', 's'},
			isMatch:  false,
			canMatch: false,
		},
		// test alternation
		{
			query:    `a+|b+`,
			seq:      []byte{},
			isMatch:  false,
			canMatch: true,
		},
		{
			query:    `a+|b+`,
			seq:      []byte{'a'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `a+|b+`,
			seq:      []byte{'b'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `a+|b+`,
			seq:      []byte{'a', 'a'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `a+|b+`,
			seq:      []byte{'b', 'b'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `a+|b+`,
			seq:      []byte{'a', 'b'},
			isMatch:  false,
			canMatch: false,
		},
		{
			query:    `a+|b+`,
			seq:      []byte{'b', 'a'},
			isMatch:  false,
			canMatch: false,
		},
		// test others
		{
			query:    `[a-z]?[1-9]*`,
			seq:      []byte{},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `[a-z]?[1-9]*`,
			seq:      []byte{'a'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `[a-z]?[1-9]*`,
			seq:      []byte{'a', '1'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `[a-z]?[1-9]*`,
			seq:      []byte{'a', '1', '2'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `[a-z]?[1-9]*`,
			seq:      []byte{'a', '1', '2', 'z'},
			isMatch:  false,
			canMatch: false,
		},
		{
			query:    `[a-z]?[1-9]*`,
			seq:      []byte{'a', 'b'},
			isMatch:  false,
			canMatch: false,
		},
		// basic case insensitive match literals
		{
			query:    `(?i)mArTy`,
			seq:      []byte{'m', 'a', 'r', 't', 'y'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `(?i)marty`,
			seq:      []byte{'m', 'A', 'r', 'T', 'y'},
			isMatch:  true,
			canMatch: true,
		},
		// case insensitive character class
		{
			query:    `(?i)[d-f]*`,
			seq:      []byte{'D', 'e', 'e', 'F'},
			isMatch:  true,
			canMatch: true,
		},
		// case insensitive, with case sensitive pattern in the middle
		{
			query:    `(?i)[d-f]*(?-i:m)wow`,
			seq:      []byte{'D', 'e', 'e', 'F', 'm', 'W', 'o', 'W'},
			isMatch:  true,
			canMatch: true,
		},
		// (?i)caseless(?-i)cased
		{
			query:    `(?i)[d-f]*(?-i)wOw`,
			seq:      []byte{'D', 'e', 'e', 'F', 'w', 'O', 'w'},
			isMatch:  true,
			canMatch: true,
		},
		// from: https://docs.rs/crate/regex-syntax/0.2.4/source/src/lib.rs
		// `(?i)[^x]` really should match any character sans `x` and `X`, but if `[^x]` is negated
		// before being case folded, you'll end up matching any character.
		{
			query:    `(?i)[^x]`,
			seq:      []byte{'a'},
			isMatch:  true,
			canMatch: true,
		},
		{
			query:    `(?i)[^x]`,
			seq:      []byte{'x'},
			isMatch:  false,
			canMatch: false,
		},
		{
			query:    `(?i)[^x]`,
			seq:      []byte{'X'},
			isMatch:  false,
			canMatch: false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s - %v", test.query, test.seq), func(t *testing.T) {
			r, err := New(test.query)
			if err != nil {
				t.Fatal(err)
			}

			s := r.Start()
			for _, b := range test.seq {
				s = r.Accept(s, b)
			}

			isMatch := r.IsMatch(s)
			if isMatch != test.isMatch {
				t.Errorf("expected isMatch %t, got %t", test.isMatch, isMatch)
			}

			canMatch := r.CanMatch(s)
			if canMatch != test.canMatch {
				t.Errorf("expectec canMatch %t, got %t", test.canMatch, canMatch)
			}
		})
	}

}

func BenchmarkNewWildcard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New("my.*h")
	}
}
