package service

import (
	"errors"
)

type Method int
type Status int
type State int
type Phase int
type Action int
type FailureType int
type MessageType int

var (
	ErrUnknown          = errors.New("unknown error")
	ErrServiceCrash     = errors.New("the service crashed")
	ErrLinkBroken       = errors.New("the communication link breaks")
	ErrTimeout          = errors.New("the communication timeout")
	ErrTTLExpired       = errors.New("ttl has expired")
	ErrUnrecoverable    = errors.New("unrecoverable error")
	ErrTooManyRetries   = errors.New("too many retires")
	ErrNoNmoreService   = errors.New("no more service")
	ErrWrongMessageType = errors.New("wrong message type")
	ErrWrongEndpoint    = errors.New("wrong endpoint")
	ErrWrongStage       = errors.New("wrong stage")

	ErrMissingOrderID    = errors.New("missing order id")
	ErrMissingCusomterID = errors.New("missing customer id")

	ErrEmptyQueue = errors.New("empty queue")
)

const (
	DefaultRetryTime = 5
)

const (
	ServiceGateway      = "gateway"
	ServiceTxManager    = "tx_manager"
	ServiceEventQueue   = "event_queue"
	ServicePayment      = "payment"
	ServiceOrder        = "order"
	ServiceShipping     = "shipping"
	ServiceCustomer     = "customer"
	ServiceNotification = "notification"
)

const (
	PhaseBegin Phase = iota + 1
	PhaseProcessing
	PhaseRollback
	PhaseEnd
)

const (
	StateNone State = iota
	StateInProgress
	StateCommit
	StateAbort
	StateComplete
)

const (
	ActionNone Action = iota
	ActionCheckpoint
)

const (
	FailureNone FailureType = iota
	FailureCrash
	FailureLinkBroken
)

const (
	MethodGet Method = iota + 1
)

const (
	StatusOK Status = iota
	StatusFailed
)

type Request struct {
	TxID     string
	Service  string
	Endpoint string
	Body     map[string]interface{}
}

type Service interface {
	Name() string
	Receive()
}

type FakeService struct {
}

func NewFakeService() Service {
	return &FakeService{}
}

func (fs *FakeService) Name() string {
	return "fake"
}

func (fs *FakeService) Receive() {
}
