package reviewerselector

import (
	"sync"
	"time"

	rand "math/rand/v2"

	"github.com/google/uuid"

	"avito-test-pr-service/internal/domain/services"
)

var _ services.ReviewerSelector = (*RandomReviewerSelector)(nil)

type RandomReviewerSelector struct {
	rnd *rand.Rand
	mu  sync.Mutex
}

func NewRandomReviewerSelector() services.ReviewerSelector {
	seed := uint64(time.Now().UnixNano())
	return NewRandomReviewerSelectorWithRand(rand.New(rand.NewPCG(seed, seed>>1|1)))
}

func NewRandomReviewerSelectorWithRand(r *rand.Rand) services.ReviewerSelector {
	if r == nil {
		seed := uint64(time.Now().UnixNano())
		r = rand.New(rand.NewPCG(seed, seed>>1|1))
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
		defer s.mu.Unlock()
		s.rnd.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

	}

	if count > len(shuffled) {
		count = len(shuffled)
	}

	return append([]uuid.UUID(nil), shuffled[:count]...)
}
