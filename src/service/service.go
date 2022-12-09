package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
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
)

const (
	DefaultRetryTime = 5
)

const (
	MethodGet Method = iota + 1
)

const (
	StatusOK Status = iota
	StatusFailed
)

const (
	PhasePrepare Phase = iota + 1
	PhaseStart
	PhaseProcessing
	PhaseEnd
)

const (
	ServiceGateway      = "gateway"
	ServiceExecutor     = "executor"
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

const (
	TypeRequest MessageType = iota
	TypeResponse
)

const (
	StateNone State = iota
	StateCommit
	StateAbort
)

const (
	ActionNone Action = iota
	ActionCheckpoint
)

type Request struct {
	ServiceName string
}

type Message struct {
	TxID               string
	Service            string
	NextService        string
	StartRound         int
	CurrentRound       int
	EndRound           int
	CurrentRetryTime   int
	RemainingRetryTime int
	Endpoint           string
	Stage              int
	MessageType        MessageType
	Phase              Phase
	State              State
	Action             Action
	Controller         string
	Stack              []string
	Body               map[string]interface{}
}

func NewMessage() Message {
	return Message{
		Stack: []string{},
		Body:  map[string]interface{}{},
	}
}

func LogMessage(msg *Message) {
	fmt.Printf("[%d] TxID: %s %s -> %s Retry %d time(s)\n", msg.CurrentRound, msg.TxID, msg.Service, msg.NextService, msg.CurrentRetryTime)
}

func LogErrorMessage(msg *Message, err error) {
	fmt.Printf("[%d] TxID: %s %s -> %s Retry %d time(s). Err: %v\n", msg.CurrentRound, msg.TxID, msg.Service, msg.NextService, msg.CurrentRetryTime, err)
}

func LogDoneMessage(msg *Message) {
	fmt.Printf("[%d] TxID: %s is done!\n", msg.CurrentRound, msg.TxID)
}

func (m *Message) IsControlledBy(s string) bool {
	return m.Controller == s
}

func (m *Message) Commit(action Action) {
	m.State = StateCommit
	m.Action = action
}

func (m *Message) Abort() {
	if m.State == StateNone {
		m.State = StateAbort
	}
}

func (m *Message) PopStack() (string, bool) {
	if len(m.Stack) == 0 {
		return "", false
	}
	lastIndex := len(m.Stack) - 1
	srv := m.Stack[lastIndex]
	m.Stack = m.Stack[:lastIndex]
	return srv, true
}

func (m *Message) PushStack(srv, endpoint string, stage int) {
	m.Stack = append(m.Stack, fmt.Sprintf("%s|%s|%d", srv, endpoint, stage))
}

func (m *Message) Get(s string) (interface{}, bool) {
	v, ok := m.Body[s]
	return v, ok
}

func (m *Message) Set(s string, v interface{}) {
	m.Body[s] = v
}

func (m *Message) Delete(s string) {
	delete(m.Body, s)
}

// dest = Service|Endpoint|Stage
func ParseDestination(dest string) (string, string, string) {
	l := strings.Split(dest, "|")
	return l[0], l[1], l[2]
}

func NextRetryRound(currentRound, retryTime int) int {
	if retryTime <= 0 {
		return -1
	}
	if retryTime > 5 {
		retryTime = 5
	}
	return currentRound + int(math.Exp2(float64(retryTime)))
}

func HandleRetryMessage(msg Message, err error, srv Service) {
	switch err {
	case ErrUnknown, ErrUnrecoverable:
		LogErrorMessage(&msg, ErrUnrecoverable)
	default:
		msg.CurrentRetryTime++
		msg.RemainingRetryTime--
		if msg.RemainingRetryTime != 0 {
			srv.Send(msg, NextRetryRound(msg.CurrentRound, msg.CurrentRetryTime))
		} else {
			LogErrorMessage(&msg, ErrTooManyRetries)
		}
	}
}

type Service interface {
	Name() string
	Send(Message, int)
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

func (fs *FakeService) Send(m Message, round int) {
}

func (fs *FakeService) Receive() {
}
