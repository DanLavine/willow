module github.com/DanLavine/willow

go 1.19

require (
	github.com/DanLavine/goasync v0.0.0-20221204033030-6fbe23639bea
	github.com/DanLavine/gomultiplex v0.0.0-20221115024652-cf8402baf491
	github.com/DanLavine/gonotify v0.0.0-20221208064238-53188eaa5cf4
	github.com/DanLavine/willow-message v0.0.0-20221116002806-eaefd527f451
	go.uber.org/zap v1.24.0
	golang.org/x/net v0.4.0
)

require (
	github.com/DanLavine/gothreadsafebuffer v0.0.0-20221211055608-f4adaa27f6a9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/text v0.5.0 // indirect
)

replace (
	github.com/DanLavine/gomultiplex => ../gomultiplex
	github.com/DanLavine/willow-message => ../willow-message
)
