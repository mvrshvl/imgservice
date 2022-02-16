package linking

type Prices struct {
	currency map[string]Price
}

// Обновить использование таблицы аккаунтов
// переписать связывание на сущности -> создание нового узла (single account uuid/cluster uuid), храним адреса сущности в мапе
// Написать получение цен на контракты
// Написать подсчет рейтинга для нод

type Price struct {
	price  float64
	amount uint64
}
