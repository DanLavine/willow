package lockmanager

import (
	"context"
	"sync/atomic"
)

type exclusiveClaim struct {
	claim                  chan struct{}
	clientsWaitingForClaim *atomic.Uint64

	destroy chan struct{} // claim was destroyed because there are no more clients waiting
	release chan struct{} // claim was released
}

func newExclusiveClaim() *exclusiveClaim {
	clientsWaiting := new(atomic.Uint64)

	return &exclusiveClaim{
		claim:                  make(chan struct{}),
		clientsWaitingForClaim: clientsWaiting,
		destroy:                make(chan struct{}),
		release:                make(chan struct{}),
	}
}

func (exclusiveClaim *exclusiveClaim) Execute(ctx context.Context) error {
	defer func() {
		close(exclusiveClaim.release)
		close(exclusiveClaim.claim)
	}()

	for {
		select {
		// service was told to shutdown
		case <-ctx.Done():
			return nil

		// exclusive claim was destroyed
		case <-exclusiveClaim.destroy:
			return nil

		// a client claimed the notify message
		case exclusiveClaim.claim <- struct{}{}:
			select {
			// service was told to shutdown
			case <-ctx.Done():
				return nil

			// the claim was released, can allow anoter process to grab the claim
			case exclusiveClaim.release <- struct{}{}:
			}
		}
	}
}

func (exclusiveClaim *exclusiveClaim) addClientWaiting() {
	exclusiveClaim.clientsWaitingForClaim.Add(1)
}

func (exclusiveClaim *exclusiveClaim) removeClientWaiting(release bool) uint64 {
	// if there are no more clients, then remove the claim
	clientsWaiting := exclusiveClaim.clientsWaitingForClaim.Add(^uint64(0))

	// need to release the claim from the current client
	if release {
		<-exclusiveClaim.release
	}

	if clientsWaiting == 0 {
		// destroy the lock
		close(exclusiveClaim.destroy)

		// wait for the destroy to finish
		<-exclusiveClaim.release
	}

	return clientsWaiting
}
