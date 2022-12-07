package brokers

import (
	"fmt"

	"github.com/DanLavine/gomultiplex"
	"github.com/DanLavine/willow-message/protocol"
	"github.com/DanLavine/willow/pkg/brokers/v1brokers"
	"github.com/DanLavine/willow/pkg/multiplex"
	"go.uber.org/zap"
)

type BrokerManager interface {
	Metrics() Metrics
	HandleConnection(logger *zap.Logger, pipe *gomultiplex.Pipe)
}

type brokerManager struct {
	// versioned brokers
	v1Broker v1brokers.BrokerManager
}

func NewBrokerManager(v1brokerManager v1brokers.BrokerManager) *brokerManager {
	return &brokerManager{
		v1Broker: v1brokerManager,
	}
}

// Handle a new connection received by the TCP server. This flters on the Header
// version and needs to be either a "Create" or "Connect" request to a broker
//
// No error is reaturned, but a malformed or incorrect request will write an
// error back to the client and close the Pipe
func (bm *brokerManager) HandleConnection(logger *zap.Logger, pipe *gomultiplex.Pipe) {
	logger = logger.Named("HandlePipe").With(zap.Uint32("pipe_id", pipe.ID()))

	// read the original header
	header := protocol.NewHeader()
	if _, err := pipe.Read(header); err != nil {
		logger.Error("failed reading header", zap.Error(err))
		multiplex.WriteError(logger, pipe, 400, err.Error())
		return
	}

	// validate the header
	if err := header.Validate(); err != nil {
		logger.Error("failed validating header", zap.Error(err))
		multiplex.WriteError(logger, pipe, 400, err.Error())
		return
	}

	switch version := header.Version(); version {
	case 1:
		bm.v1Broker.AddPipe(logger, pipe)
	default:
		logger.Error("Invalid version received", zap.Uint32("version", version))
		multiplex.WriteError(logger, pipe, 400, fmt.Sprintf("Invalid header version '%d'", version))
		return
	}
}

type Metrics struct {
	V1BrokerInfo v1brokers.BrokerMetrics
}

func (bm *brokerManager) Metrics() Metrics {
	return Metrics{
		V1BrokerInfo: bm.v1Broker.Metrics(),
	}
}
