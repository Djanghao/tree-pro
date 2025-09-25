package internal

import "fmt"

// DirGroup represents a collection of directories that share the same structure signature.
type DirGroup struct {
	Signature string
	Members   []*Directory
}

// GroupIdentical partitions directories into groups of identical structures while preserving
// their original order of appearance.
func GroupIdentical(dirs []*Directory) []DirGroup {
	grouped := make(map[string][]*Directory)
	order := make([]string, 0)

	for _, dir := range dirs {
		if dir == nil {
			continue
		}
		sig := dir.Signature
		if sig == "" {
			sig = fallbackSignature(dir)
		}
		if _, seen := grouped[sig]; !seen {
			order = append(order, sig)
		}
		grouped[sig] = append(grouped[sig], dir)
	}

	result := make([]DirGroup, 0, len(order))
	for _, sig := range order {
		result = append(result, DirGroup{
			Signature: sig,
			Members:   grouped[sig],
		})
	}

	return result
}

func fallbackSignature(dir *Directory) string {
	return fmt.Sprintf("name:%s:level:%d", dir.Name, dir.Level)
}
