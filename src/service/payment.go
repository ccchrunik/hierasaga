package service

import "atm/ds"

type PaymentEntry struct {
	TxID       string
	PaymentID  string
	OrderID    string
	CustomerID string
}

type PaymentDB struct {
	entries map[string]PaymentEntry
}

func NewPaymentDB() *PaymentDB {
	return &PaymentDB{
		entries: map[string]PaymentEntry{},
	}
}

type PaymentService struct {
	cfg   *SystemConfig
	queue ds.Queue
	db    *PaymentDB
}

func NewPaymentService(cfg *SystemConfig) *PaymentService {
	return &PaymentService{
		cfg:   cfg,
		queue: ds.NewMutexTimedPriorityQueue(&cfg.round),
		db:    NewPaymentDB(),
	}
}

func (pdb *PaymentDB) createNewPayment(txID string, orderID string, customerID string) error {
	return nil
}

func (ps *PaymentService) Name() string {
	return ServicePayment
}

func (ps *PaymentService) Send(msg Message) {
	ps.queue.Push(ds.NewItem(ps.cfg.round+1, msg))
}

// assume only 1 instance
func (ps *PaymentService) Receive() {
	services := ps.cfg.services
	executor := services[ServiceExecutor]
	nextRound := ps.cfg.round + 1
	for !ps.queue.IsEmpty() {
		msg := ps.queue.Pop().(Message)
		LogMessage(&msg)
		switch msg.Endpoint {
		// create a new payment (control endpoint)
		case "new_payment":
			switch msg.MessageType {
			case TypeRequest:
				switch msg.Stage {
				case 1:
					msg.PushStack(ServicePayment, "new_payment", 2)
					msg.Service = ServicePayment
					msg.NextService = ServicePayment
					msg.Endpoint = "insert_payment"
					msg.Stage = 1
					executor.Send(msg, nextRound)

				case 2:
					// update order status
					msg.PushStack(ServicePayment, "new_payment", 3)
					msg.Service = ServicePayment
					msg.NextService = ServiceOrder
					msg.Endpoint = "update_order"
					msg.Stage = 1
					executor.Send(msg, nextRound)

				case 3:
					// update customer information (e.g. add shopping points)
					msg.PushStack(ServicePayment, "new_payment", 4)
					msg.Service = ServicePayment
					msg.NextService = ServiceCustomer
					msg.Endpoint = "update_customer"
					msg.Stage = 1
					executor.Send(msg, nextRound)

				case 4:
					// mark this transaction as completed
					msg.Phase = PhaseEnd
					executor.Send(msg, nextRound)

				default:
					LogErrorMessage(&msg, ErrWrongStage)
				}
			default:
				LogErrorMessage(&msg, ErrWrongMessageType)
			}

		case "insert_payment":
			v, ok := msg.Get("OrderID")
			if !ok {
				LogErrorMessage(&msg, ErrMissingOrderID)
			}
			orderID := v.(string)
			v, ok = msg.Get("CustomerID")
			if !ok {
				LogErrorMessage(&msg, ErrMissingCusomterID)
			}
			customerID := v.(string)

			ps.db.createNewPayment(msg.TxID, orderID, customerID)
			executor.Send(msg, nextRound)

		default:
			LogErrorMessage(&msg, ErrWrongEndpoint)
		}
	}
}
