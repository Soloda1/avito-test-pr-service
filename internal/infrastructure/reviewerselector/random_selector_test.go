package reviewerselector

import (
	rand "math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomReviewerSelector_Select(t *testing.T) {
	candidates := []string{"id-1", "id-2", "id-3", "id-4"}

	tests := []struct {
		name       string
		candidates []string
		count      int
		seed       uint64
		expectLen  int
	}{
		{"select subset", candidates, 2, 1, 2},
		{"count greater than candidates", candidates[:2], 5, 42, 2},
		{"zero count returns nil", candidates, 0, 99, 0},
		{"no candidates", nil, 2, 7, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRandomReviewerSelectorWithRand(rand.New(rand.NewPCG(tt.seed, tt.seed>>1|1)))
			result := selector.Select(tt.candidates, tt.count)

			if tt.expectLen == 0 && tt.count > 0 && len(tt.candidates) == 0 {
				require.Nil(t, result)
				return
			}
			if tt.expectLen == 0 && tt.count == 0 {
				require.Nil(t, result)
				return
			}
			require.Len(t, result, tt.expectLen)
			seen := make(map[string]struct{}, len(result))
			for _, id := range result {
				_, exists := seen[id]
				require.False(t, exists, "duplicate reviewer returned")
				seen[id] = struct{}{}
			}
			for id := range seen {
				found := false
				for _, c := range tt.candidates {
					if c == id {
						found = true
						break
					}
				}
				require.True(t, found, "selected id not in candidates")
			}
		})
	}
}

func TestNewRandomReviewerSelectorWithRand_NilFallback(t *testing.T) {
	selector := NewRandomReviewerSelectorWithRand(nil)
	result := selector.Select([]string{"x"}, 1)
	require.Len(t, result, 1)
}
