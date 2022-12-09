package service

type CustomerEntry struct {
	TxID       string
	CustomerID string
}

type CustomerDB struct {
	entries map[string]CustomerEntry
}

type CustomerService struct {
}
