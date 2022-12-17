package v1queues

import v1 "github.com/DanLavine/willow-message/protocol/v1"

type Message struct {
	v1Message *v1.Message
}
