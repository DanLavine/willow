Reading and writing to Disk
-------------

Writing to disk for all the services is another massive operation that will take some experimentation to
figure out the exact requirements. The initial ones that come to mind are:

1. Processes can save their state to disk and restart successfully
2. Processes when saved to state no longer need to stay in memory if they can help it.
3. Background heartbeat operations need to handle restarts for remote clients and not have them time out

# Locker

Take the simplest service, Locker for example. When a lock is created it can at least save the state of the
lock to persist for a restart and have clients continue to heartbeat for the lock they already have. But in addition
to this, does the lock have to stay in memory for the BTree? Once there is already a lock, what would a second
request do when trying to load this Lock? Does the first request have to finish processing or time out? Could depend
on how we want the clients to grab the lock. I.E round robbin, fifo, etc

# Limiter

Limiter on the other hand can split some of the use cases for writing to disk. The Counters and Rules that are created
don't need to stay in memory and can be looked up on demand when they are needed. So the BTrees clearly need to support
these operations to unload the tree structures and read the from disk when needed and reconstruct the Limiters objects when
needed.

# Willow

Willow is a bit of a split camp where Clients wanting to dequeue items need to have them loaded in memory, but if a queue
has no Clients waiting for them, then those objects could be unloaded and simply update their Queue Items when enqueue
requests come in.