package multiplex

import (
	"github.com/DanLavine/gomultiplex"
	v1 "github.com/DanLavine/willow-message/protocol/v1"
	"go.uber.org/zap"
)

// Write an error to the client and close the pipe after sending the message
func WriteError(logger *zap.Logger, pipe *gomultiplex.Pipe, statusCode int, errReason string) {
	logger = logger.Named("WriteError")

	body, err := v1.NewError(200, errReason)
	if err == nil {
		// don't care about these errors. The pipe is going to be closed so if the client isn't receiving messages
		// then thats fine, the close should case a log on their side
		_, _ = pipe.Write(body)
	} else {
		logger.Error("failed creaing error message", zap.Error(err))
	}

	pipe.Close()
}
