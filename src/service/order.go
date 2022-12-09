package service

type Product struct {
	ProductID string
	Name      string
	Price     float64
}

type OrderEntry struct {
	TxID     string
	OrderID  string
	Buyer    string
	Seller   string
	Amount   string
	Products []string
}

type OrderDB struct {
	entries map[string]OrderEntry
}

type OrderService struct {
}
