module github.com/DanLavine/willow

go 1.19

require (
	github.com/DanLavine/goasync v0.0.0-20220905181031-79c296fed01e
	github.com/DanLavine/gomultiplex v0.0.0-20221115024652-cf8402baf491
	github.com/DanLavine/willow-message v0.0.0-20221116002806-eaefd527f451
	go.uber.org/zap v1.23.0
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)

replace (
	github.com/DanLavine/gomultiplex => ../gomultiplex
	github.com/DanLavine/willow-message => ../willow-message
)
