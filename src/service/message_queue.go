package service

import "atm/ds"

type MessageQueue struct {
	queue *ds.MutexQueue
}

// func NewMessageQueue(cfg *SystemConfig) *MessageQueue {
// 	return &MessageQueue{
// 		queue: ds.NewMutexQueue(),
// 	}
// }
