package reviewerselector

import (
	"sync"
	"time"

	rand "math/rand/v2"

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

func (s *RandomReviewerSelector) Select(candidates []string, count int) []string {
	if count <= 0 || len(candidates) == 0 {
		return nil
	}

	shuffled := append([]string(nil), candidates...)
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

	return append([]string(nil), shuffled[:count]...)
}
