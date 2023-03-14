package tags

import v1 "github.com/DanLavine/willow/pkg/models/v1"

type Tag interface {
	Process() *v1.DequeueItemResponse
}
