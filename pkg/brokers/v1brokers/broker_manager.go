package v1brokers

import (
	"encoding/json"

	"github.com/DanLavine/gomultiplex"
	"github.com/DanLavine/willow-message/protocol"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type BrokerManager interface {
	AddPipe(logger *zap.Logger, pipe *gomultiplex.Pipe)
}

type brokerManager struct {
	queueManager QueueManager
	//pubSubManager PubSubManager
}

func NewBrokerManager(queueManager QueueManager) *brokerManager {
	return &brokerManager{
		queueManager: queueManager,
	}
}

// Add pipe can be used to setup a new connection to the broker
func (bm *brokerManager) AddPipe(logger *zap.Logger, pipe *gomultiplex.Pipe) {
	logger = logger.Named("AddPipe")

	// read the v1 header
	header := v1.NewHeader()
	if _, err := pipe.Read(header); err != nil {
		logger.Error("failed reading v1 header", zap.Error(err))
		multiplex.WriteError(logger, pipe, 400, err.Error())
		return
	}

	// validate the header
	if err := header.Validate(); err != nil {
		logger.Error("failed validating header", zap.Error(err))
		multiplex.WriteError(logger, pipe, 400, err.Error())
		return
	}

	// read the body
	body := make([]byte, header.BodySize())
	if _, err := pipe.Read(body); err != nil {
		logger.Error("failed reading message body", zap.Error(err))
		multiplex.WriteError(logger, pipe, 500, err.Error())
		return
	}

	switch header.MessageType() {
	case v1.TypeCreate:
		// parse create request
		createRequest := &v1.Create{}
		if err := json.Unmarshal(body, createRequest); err != nil {
			logger.Error("Failed parsing create request", zap.Error(err))
			multiplex.WriteError(logger, pipe, 500, "Internal Server Error")
			return
		}

		// create either the QUEUE or PUB-SUB brokers
		switch brokerType := createRequest.Broker; brokerType {
		case protocol.Queue:
			bm.queueManager.CreateChannel(logger, createRequest, pipe)
		case protocol.PubSub:
			logger.Error("Failed create PUBSUB request. Unimplemented", zap.Any("protocol", brokerType))
			multiplex.WriteError(logger, pipe, 501, "Unimplemented create PubSub")
			return
		default:
			logger.Error("Failed create request. Received unkown protocol", zap.Any("protocol", brokerType))
			multiplex.WriteError(logger, pipe, 400, "Unexpected Broker Type received. Must be [Queue | PubSub]")
			return
		}
	case v1.TypeConnect:
		// parse the connect request
		connectRequest := &v1.Connect{}
		if err := json.Unmarshal(body, connectRequest); err != nil {
			logger.Error("Failed parsing connect request", zap.Error(err))
			multiplex.WriteError(logger, pipe, 500, "Internal Server Error")
			return
		}

		// connect to either the QUEUE or PUB-SUB brokers
		switch brokerType := connectRequest.Broker; brokerType {
		case protocol.Queue:
			bm.queueManager.ConnectChannel(logger, connectRequest, pipe)
		case protocol.PubSub:
			logger.Error("Failed create PUBSUB request. Unimplemented", zap.Any("protocol", brokerType))
			return
		default:
			logger.Error("Failed create request. Received unkown protocol", zap.Any("protocol", brokerType))
			multiplex.WriteError(logger, pipe, 400, "Unexpected Broker Type received. Must be [Queue | PubSub]")
			return
		}
	default:
		logger.Error("Unexpected initial header", zap.Uint32("header", uint32(header.MessageType())))
		multiplex.WriteError(logger, pipe, 400, "Unexpected header. Must be [Create | Connect]")
		return
	}
}
