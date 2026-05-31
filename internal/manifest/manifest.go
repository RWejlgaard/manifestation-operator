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

// Package manifest holds the shared vocabulary of the manifestation operator: the
// scheduling gate that keeps pods in limbo, the namespace opt-in, and the rules the
// universe uses to decide whether your affirmation is worthy.
package manifest

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// SchedulingGate is the scheduling gate injected onto pods. While present, the
	// scheduler refuses to bind the pod and it sits in Pending/SchedulingGated. It is
	// removed only when a worthy Desire manifests the pod.
	SchedulingGate = "manifestation.pez.sh/awaiting-manifestation"

	// EnabledLabel opts a namespace into manifestation. Without it, the webhook keeps
	// its hands off and your pods schedule like in a sad, materialist universe.
	EnabledLabel = "manifestation.pez.sh/enabled"

	// ManifestedByAnnotation records which Desire released a given pod from limbo.
	ManifestedByAnnotation = "manifestation.pez.sh/manifested-by"

	// SkipLabel lets a single pod opt out even in an enabled namespace.
	SkipLabel = "manifestation.pez.sh/skip"
)

// doubtWords betray a lack of faith: future tense, conditionals, hope, wishing. The
// universe does not respond to any of these.
var doubtWords = []string{
	"will", "shall", "gonna", "going to", "want to", "wanna", "wish", "hope",
	"would", "could", "should", "might", "maybe", "perhaps", "soon", "eventually",
	"one day", "someday", "trying to", "try to", "need to", "needs to", "let's",
	"please", "if ", "fix", "todo", "tomorrow", "later",
}

// pastWords betray a fixation on what already happened. Manifest the now, not the then.
var pastWords = []string{
	" was ", " were ", " had ", " did ", "used to", " has been ", " have been ",
}

// presentCopula are the present-tense verbs of being. At least one must appear: the
// universe needs you to declare that the thing simply IS.
var presentCopula = regexp.MustCompile(`(?i)\b(is|are|am|runs?|serves?|works?|flows?|thrives?|stays?|remains?|holds?|stands?|lives?|breathes?)\b|('s|'re|'m)\b`)

// Validate decides whether an affirmation is worthy of the universe. It returns ok and
// a reason that is either congratulation or gentle cosmic correction.
func Validate(manifestation string) (ok bool, reason string) {
	m := strings.TrimSpace(manifestation)
	lower := strings.ToLower(m)

	if m == "" {
		return false, "The void hears only silence. Speak your desire in present tense."
	}
	if strings.Contains(m, "?") {
		return false, "A question is doubt wearing punctuation. State it as already true."
	}
	if strings.HasSuffix(m, "!") && strings.Count(m, "!") > 3 {
		return false, "Easy, achi. The universe hears you. Lose the desperation."
	}

	for _, w := range doubtWords {
		if containsWord(lower, w) {
			return false, fmt.Sprintf("The universe heard '%s' and felt your doubt. Speak as if it already is.", strings.TrimSpace(w))
		}
	}
	for _, w := range pastWords {
		if strings.Contains(lower, w) {
			return false, fmt.Sprintf("'%s' lives in the past. Manifest the present.", strings.TrimSpace(w))
		}
	}

	if !presentCopula.MatchString(lower) {
		return false, "This is not present tense. Tell the universe what IS, right now."
	}

	return true, "The universe accepts your truth. It is already so."
}

// containsWord matches a word with simple boundaries so "will" does not trip on "skillful".
func containsWord(haystack, needle string) bool {
	needle = strings.TrimSpace(needle)
	if needle == "" {
		return false
	}
	// Multi-word phrases: substring match is fine and intended.
	if strings.Contains(needle, " ") {
		return strings.Contains(haystack, needle)
	}
	for _, f := range strings.FieldsFunc(haystack, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	}) {
		if f == needle {
			return true
		}
	}
	return false
}
