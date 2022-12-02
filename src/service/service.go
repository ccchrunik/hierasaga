package service

type Method int
type Status int
type FailureType int

const (
	MethodGet Method = iota + 1
)

const (
	StatusOk Status = iota + 1
)

const (
	ServiceGateway      = "gateway"
	ServiceTxExecutor   = "tx_executor"
	ServiceTxManager    = "tx_manager"
	ServiceMessageQueue = "message_queue"
	ServiceCustomer     = "customer"
	ServiceOrder        = "order"
	ServicePayment      = "payment"
)

const (
	FailureNone FailureType = iota
	FailureCrash
	FailureLinkBroken
)

type Request struct {
	ServiceName string
}

type Message struct {
	TxID             string
	ServiceFrom      string
	ServiceTo        string
	StartRound       int
	EndRound         int
	CurrentRetryTime int
	Body             map[string]string
}

type Service interface {
	Name() string
	Setup() error
	Send(Message) error
	Execute() error
}
