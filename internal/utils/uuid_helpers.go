package utils

import "github.com/google/uuid"

func FilterUUIDs(candidates []uuid.UUID, exclude map[uuid.UUID]struct{}) []uuid.UUID {
	if len(candidates) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(candidates))
	for _, id := range candidates {
		if _, ok := exclude[id]; ok {
			continue
		}
		out = append(out, id)
	}
	return out
}

func UniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) <= 1 {
		return ids
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func ContainsUUID(list []uuid.UUID, id uuid.UUID) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}
