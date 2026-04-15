package interactive

import "sort"

// Group represents a named group of items.
type Group[T any] struct {
	Label string
	Items []T
}

// GroupBy groups items using the provided key function.
// Groups are returned sorted by label.
func GroupBy[T any](items []T, keyFn func(T) string) []Group[T] {
	idx := make(map[string]int)
	var groups []Group[T]

	for _, item := range items {
		key := keyFn(item)
		if i, ok := idx[key]; ok {
			groups[i].Items = append(groups[i].Items, item)
		} else {
			idx[key] = len(groups)
			groups = append(groups, Group[T]{Label: key, Items: []T{item}})
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Label < groups[j].Label
	})

	return groups
}
