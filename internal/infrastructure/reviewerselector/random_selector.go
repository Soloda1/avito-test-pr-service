package reviewerselector

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"

	"avito-test-pr-service/internal/domain/services"
)

var _ services.ReviewerSelector = (*RandomReviewerSelector)(nil)

type RandomReviewerSelector struct {
	rnd *rand.Rand
	mu  sync.Mutex
}

func NewRandomReviewerSelector() services.ReviewerSelector {
	return NewRandomReviewerSelectorWithRand(rand.New(rand.NewSource(time.Now().UnixNano())))
}

func NewRandomReviewerSelectorWithRand(r *rand.Rand) services.ReviewerSelector {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &RandomReviewerSelector{rnd: r}
}

func (s *RandomReviewerSelector) Select(candidates []uuid.UUID, count int) []uuid.UUID {
	if count <= 0 || len(candidates) == 0 {
		return nil
	}

	shuffled := append([]uuid.UUID(nil), candidates...)
	if len(shuffled) > 1 {
		s.mu.Lock()
		s.rnd.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		s.mu.Unlock()
	}

	if count > len(shuffled) {
		count = len(shuffled)
	}

	return append([]uuid.UUID(nil), shuffled[:count]...)
}
