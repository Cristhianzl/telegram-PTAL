package engine

import (
	"fmt"
	"sort"
	"strings"
)

// AllKinds lists every event the engine can produce, in the order they are
// shown to the user. Keeping this beside Kind means a new event type is
// configurable the moment it exists, rather than silently unfilterable.
var AllKinds = []Kind{
	KindReviewRequested,
	KindMentioned,
	KindAssigned,
	KindChangesRequested,
	KindApproved,
	KindChecksFailed,
	KindChecksFixed,
	KindConflict,
	KindNewActivity,
	KindReadyForReview,
	KindGone,
}

// Summary explains, in one line, what triggers this event.
func (k Kind) Summary() string {
	switch k {
	case KindReviewRequested:
		return "Someone asked you to review a pull request"
	case KindMentioned:
		return "Someone @-mentioned you"
	case KindAssigned:
		return "A pull request was assigned to you"
	case KindChangesRequested:
		return "A reviewer requested changes on your pull request"
	case KindApproved:
		return "Your pull request was approved"
	case KindChecksFailed:
		return "CI failed on your pull request"
	case KindChecksFixed:
		return "CI recovered on your pull request"
	case KindConflict:
		return "Your pull request developed a merge conflict"
	case KindNewActivity:
		return "New comments or reviews on a pull request of yours"
	case KindReadyForReview:
		return "A draft you are reviewing became ready"
	case KindGone:
		return "A pull request was closed or merged"
	}
	return string(k)
}

// ParseKinds turns a comma-separated list into event kinds.
//
// An unknown name is an error rather than a silent no-op: a typo in a mute
// list would otherwise leave the user believing they had turned an alert off,
// and they would only find out by being interrupted by it.
func ParseKinds(raw string) ([]Kind, error) {
	var out []Kind
	for _, name := range strings.Split(raw, ",") {
		name = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(name, "-", "_")))
		if name == "" {
			continue
		}
		kind, ok := lookupKind(name)
		if !ok {
			return nil, fmt.Errorf("unknown event %q\n  Valid events: %s",
				name, strings.Join(KindNames(), ", "))
		}
		out = append(out, kind)
	}
	return out, nil
}

func lookupKind(name string) (Kind, bool) {
	for _, k := range AllKinds {
		if string(k) == name {
			return k, true
		}
	}
	return "", false
}

// KindNames lists every event name, sorted, for help text and errors.
func KindNames() []string {
	out := make([]string, 0, len(AllKinds))
	for _, k := range AllKinds {
		out = append(out, string(k))
	}
	sort.Strings(out)
	return out
}

// KindFilter decides which events are allowed through.
type KindFilter struct {
	// allow, when non-empty, is the only set that passes.
	allow map[Kind]bool
	// mute is subtracted afterwards, and wins over allow.
	mute map[Kind]bool
}

// NewKindFilter builds a filter from the two configured lists.
//
// The two compose in the obvious order: an empty allow list means everything
// is allowed, and mute always subtracts. That way someone can either name the
// handful they want, or keep the default and turn off the one that annoys
// them, without having to think about which mechanism wins.
func NewKindFilter(allow, mute []Kind) KindFilter {
	f := KindFilter{}
	if len(allow) > 0 {
		f.allow = make(map[Kind]bool, len(allow))
		for _, k := range allow {
			f.allow[k] = true
		}
	}
	if len(mute) > 0 {
		f.mute = make(map[Kind]bool, len(mute))
		for _, k := range mute {
			f.mute[k] = true
		}
	}
	return f
}

// Allows reports whether an event of this kind should be delivered.
func (f KindFilter) Allows(k Kind) bool {
	if f.mute[k] {
		return false
	}
	if f.allow == nil {
		return true
	}
	return f.allow[k]
}

// Muted lists the kinds this filter suppresses, for `doctor` and `config`.
func (f KindFilter) Muted() []Kind {
	var out []Kind
	for _, k := range AllKinds {
		if !f.Allows(k) {
			out = append(out, k)
		}
	}
	return out
}
