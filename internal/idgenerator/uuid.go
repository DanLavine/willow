package idgenerator

import "github.com/segmentio/ksuid"

type uuid struct{}

func UUID() *uuid {
	return &uuid{}
}

func (uuid *uuid) ID() string {
	return ksuid.New().String()
}
