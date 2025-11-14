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
