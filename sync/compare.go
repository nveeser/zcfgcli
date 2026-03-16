package sync

import (
	"github.com/google/go-cmp/cmp"
	"maps"
)

func Compare(a, b *Device, onDiff func(key string, before, after any)) int {
	if onDiff == nil {
		onDiff = func(key string, before, after any) {}
	}
	unionSet := make(map[string]struct{})
	for k := range maps.Keys(a.config) {
		unionSet[k] = struct{}{}
	}
	for k := range maps.Keys(b.config) {
		unionSet[k] = struct{}{}
	}
	var diffs int
	for k := range maps.Keys(unionSet) {
		av, aOk := a.config[k]
		bv, bOk := b.config[k]

		switch {
		case !aOk && !bOk:
			panic("unionSet has invalid keys?")
		case !aOk && bOk:
			onDiff(k, nil, bv)
			diffs++
		case aOk && !bOk:
			onDiff(k, av, nil)
			diffs++
		case cmp.Diff(av, bv) != "":

			onDiff(k, av, bv)
			diffs++
		}
	}
	return diffs
}
