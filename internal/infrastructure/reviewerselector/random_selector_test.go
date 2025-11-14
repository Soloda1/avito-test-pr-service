package reviewerselector

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRandomReviewerSelector_Select(t *testing.T) {
	candidates := []uuid.UUID{
		uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		uuid.MustParse("44444444-4444-4444-4444-444444444444"),
	}

	tests := []struct {
		name        string
		candidates  []uuid.UUID
		count       int
		seed        int64
		expectNil   bool
		expectedLen int
	}{
		{
			name:        "select subset",
			candidates:  candidates,
			count:       2,
			seed:        1,
			expectedLen: 2,
		},
		{
			name:        "count greater than candidates",
			candidates:  candidates[:2],
			count:       5,
			seed:        42,
			expectedLen: 2,
		},
		{
			name:       "zero count returns nil",
			candidates: candidates,
			count:      0,
			seed:       99,
			expectNil:  true,
		},
		{
			name:       "no candidates",
			candidates: nil,
			count:      2,
			seed:       7,
			expectNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewRandomReviewerSelectorWithRand(rand.New(rand.NewSource(tt.seed)))
			result := selector.Select(tt.candidates, tt.count)

			if tt.expectNil {
				require.Nil(t, result)
				return
			}

			require.Len(t, result, tt.expectedLen)

			seen := make(map[uuid.UUID]struct{}, len(result))
			for _, id := range result {
				_, exists := seen[id]
				require.False(t, exists, "duplicate reviewer returned")
				seen[id] = struct{}{}
			}

			for id := range seen {
				require.Contains(t, tt.candidates, id)
			}
		})
	}
}

func TestNewRandomReviewerSelectorWithRand_NilFallback(t *testing.T) {
	selector := NewRandomReviewerSelectorWithRand(nil)
	result := selector.Select([]uuid.UUID{uuid.New()}, 1)
	require.Len(t, result, 1)
}
