package tags

import v1 "github.com/DanLavine/willow/pkg/models/v1"

type Tag = func() *v1.DequeueItemResponse
