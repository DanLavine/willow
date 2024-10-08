package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/urlrouter"
	"github.com/DanLavine/willow/internal/config"
	"github.com/DanLavine/willow/internal/helpers"
	"github.com/DanLavine/willow/internal/reporting"
	"github.com/DanLavine/willow/internal/willow/api"
	"github.com/DanLavine/willow/internal/willow/api/v1/handlers"
	"github.com/DanLavine/willow/internal/willow/brokers/queue_channels/constructor"
	"github.com/DanLavine/willow/internal/willow/brokers/queues"
	"github.com/DanLavine/willow/pkg/clients"
	"go.uber.org/zap"

	v1router "github.com/DanLavine/willow/internal/willow/api/v1/router"
	queuechannels "github.com/DanLavine/willow/internal/willow/brokers/queue_channels"
	limiterclient "github.com/DanLavine/willow/pkg/clients/limiter_client"
	v1 "github.com/DanLavine/willow/pkg/models/api/limiter/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func main() {
	cfg, err := config.Willow(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := reporting.NewZapLogger(cfg)
	defer logger.Sync()

	// setup shutdown signal
	shutdown, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	// setup limiterclient config and validate it
	clientConfig := &clients.Config{
		URL:           *cfg.LimiterURL,
		CAFile:        *cfg.LimiterClientCA,
		ClientKeyFile: *cfg.LimiterClientKey,
		ClientCRTFile: *cfg.LimiterClientCRT,
	}
	limiterClient, err := limiterclient.NewLimiterClient(clientConfig)
	if err != nil {
		logger.Fatal("limiter client is not configured properly", zap.Error(err))
	}

	// ensure the limiter client can connect to the locker service
	for i := 0; i < 10; i++ {
		if err := limiterClient.Healthy(); err != nil {
			logger.Error("error checking health of limiter service", zap.Error(err))
			time.Sleep(10 * time.Second)

			if i == 9 {
				logger.Fatal("Failed to setup the limiter client which is required")
			}

			continue
		}

		break
	}

	// setup the rule if it does not exist for the willow limits
	ruleResp, err := limiterClient.CreateRule(context.Background(), &v1.Rule{
		Spec: &v1.RuleSpec{
			DBDefinition: &v1.RuleDBDefinition{
				GroupByKeyValues: datatypes.KeyValues{
					"_willow_queue_name": datatypes.Any(),
					"_willow_enqueued":   datatypes.Any(),
				},
			},
			Properties: &v1.RuleProperties{
				Limit: helpers.PointerOf[int64](0), // by default, all queues have a limit of 0
			},
		},
	})
	if err != nil {
		logger.Fatal("Failed to setup Limiter enqueue rule", zap.Error(err))
	}

	// setup async handlers
	//// using strict config ensures that if any process fails, the server will ty and shutdown gracefully
	taskManager := goasync.NewTaskManager(goasync.StrictConfig())

	// queue channels client
	queueChannelsConstructor, err := constructor.NewQueueChannelConstructor("memory", limiterClient)
	if err != nil {
		logger.Fatal("Failed to setup queue channels constructor", zap.Error(err))
	}
	queueChannelsClient := queuechannels.NewLocalQueueChannelsClient(queueChannelsConstructor)
	taskManager.AddExecuteTask("queue channels client", queueChannelsClient)

	// queue client
	queueConstructor, err := queues.NewQueueConstructor("memory", limiterClient)
	if err != nil {
		logger.Fatal("Failed to setup queue constructor", zap.Error(err))
	}
	queueClient := queues.NewLocalQueueClient(queueConstructor, queueChannelsClient, ruleResp.State.ID)

	// setup willow server
	willowMux := urlrouter.New()
	//// v1 api handlers
	v1router.AddV1WillowRoutes(logger, willowMux, handlers.NewV1QueueHandler(queueClient, ruleResp.State.ID))
	taskManager.AddTask("tcp_server", api.NewWillowTCP(logger, cfg, willowMux))

	// start all processes
	if errs := taskManager.Run(shutdown); errs != nil {
		log.Fatal("Failed runnng willow cleanly", errs)
	}

	logger.Info("Successfully shutdown")
}
