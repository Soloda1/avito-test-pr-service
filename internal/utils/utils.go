package utils

// FilterStrings фильтрует список строковых идентификаторов, исключая элементы из exclude.
//
// Параметры:
//   - candidates: исходный срез строк-кандидатов
//   - exclude: множество строк для исключения (ключи map — это исключаемые значения)
//
// Возвращает:
//   - новый срез строк без исключённых значений (может быть пустой, но не nil)
func FilterStrings(candidates []string, exclude map[string]struct{}) []string {
	if len(candidates) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(candidates))
	for _, id := range candidates {
		if _, ok := exclude[id]; ok {
			continue
		}
		out = append(out, id)
	}
	return out
}

// ContainsString проверяет наличие строки в срезе.
//
// Параметры:
//   - list: срез строк, в котором выполняется поиск
//   - id: искомая строка
//
// Возвращает:
//   - true, если значение найдено; иначе false
func ContainsString(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}
