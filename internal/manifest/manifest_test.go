/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifest

import "testing"

func TestValidate(t *testing.T) {
	cases := []struct {
		name          string
		manifestation string
		want          bool
	}{
		{"present tense affirmation", "My nginx pod is healthy and serving traffic", true},
		{"present continuous", "The database is running and replication flows", true},
		{"simple is", "Everything is fine", true},
		{"contraction", "My service's healthy and it serves users", true},
		{"future will", "My pod will be healthy", false},
		{"future gonna", "It is gonna work eventually", false},
		{"wishing", "I hope the pod is healthy", false},
		{"conditional should", "The pod should be running", false},
		{"past tense", "The pod was healthy", false},
		{"a question", "Is my pod healthy?", false},
		{"empty", "", false},
		{"no copula, just a noun", "healthy nginx pod traffic", false},
		{"please betrays need", "Please make the pod healthy", false},
		{"skillful is not will", "My skillful operator is serving traffic", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := Validate(tc.manifestation)
			if got != tc.want {
				t.Errorf("Validate(%q) = %v (%q), want %v", tc.manifestation, got, reason, tc.want)
			}
		})
	}
}
