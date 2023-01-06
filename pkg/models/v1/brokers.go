package v1

type BrokerType uint32

const (
	Queue BrokerType = iota
)

func (bt BrokerType) ToString() string {
	switch bt {
	case Queue:
		return "queue"
	default:
		return "unknown"
	}
}
