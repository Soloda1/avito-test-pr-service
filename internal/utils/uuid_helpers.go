package utils

import "github.com/google/uuid"

// FilterUUIDs фильтрует список идентификаторов, исключая элементы,
// содержащиеся во множестве exclude. Порядок оставшихся элементов сохраняется.
// Всегда возвращает непустой (ненил) срез: если нет результатов, то []uuid.UUID{}.
//
// Параметры:
//   - candidates: исходный срез UUID-кандидатов
//   - exclude: множество UUID для исключения (ключи map — это исключаемые значения)
//
// Возвращает:
//   - новый срез UUID без исключённых значений (может быть пустой, но не nil)
func FilterUUIDs(candidates []uuid.UUID, exclude map[uuid.UUID]struct{}) []uuid.UUID {
	if len(candidates) == 0 {
		return []uuid.UUID{}
	}
	out := make([]uuid.UUID, 0, len(candidates))
	for _, id := range candidates {
		if _, ok := exclude[id]; ok {
			continue
		}
		out = append(out, id)
	}
	if len(out) == 0 {
		return []uuid.UUID{}
	}
	return out
}

// UniqueUUIDs убирает дубликаты из среза UUID, сохраняя первый встреченный
// экземпляр каждого значения и относительный порядок элементов.
//
// Параметры:
//   - ids: исходный срез UUID, возможно с повторами
//
// Возвращает:
//   - новый срез без дубликатов; для длины 0 или 1 возвращает исходный срез
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

// ContainsUUID выполняет линейную проверку наличия значения в срезе UUID.
//
// Параметры:
//   - list: срез UUID, в котором выполняется поиск
//   - id: искомый UUID
//
// Возвращает:
//   - true, если значение найдено; иначе false
func ContainsUUID(list []uuid.UUID, id uuid.UUID) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}
