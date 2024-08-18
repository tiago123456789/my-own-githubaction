package queue

import "encoding/json"

type IQueueUtil interface {
	ParseMessage(payload []byte, p interface{}) error
}

type QueueUtil struct {
}

func NewQueueUtil() *QueueUtil {
	return &QueueUtil{}
}

func (q *QueueUtil) ParseMessage(payload []byte, p interface{}) error {
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	return nil
}
