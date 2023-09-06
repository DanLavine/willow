module github.com/DanLavine/willow

go 1.19

require (
	github.com/DanLavine/channelops v0.0.0-20230909014756-bd208194ad96
	github.com/DanLavine/goasync v1.0.0
	github.com/DanLavine/gonotify v0.0.0-20221228000906-77ad21d2336e
	github.com/google/uuid v1.3.0
	github.com/onsi/gomega v1.27.10
	github.com/segmentio/ksuid v1.0.4
	go.uber.org/mock v0.2.0
	go.uber.org/zap v1.24.0
	golang.org/x/net v0.12.0
)

require (
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/DanLavine/channelops => ../channelops
