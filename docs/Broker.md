Broker
------

The Broker is the main API server managing Willow's message brokers. It runs a TCP connection
that Clients need to connect to.

* Port - default 8080

# Queues

Feature for Brokers:
1. Create a queue with a unique name and tags (list of strings).
  1. iff the name + tags already exist, perform a no-op success since it already exists
1. Queues can be configured with a retry count
  1. Once a message has failed == retry count. Send to the Dead Letter Queue
1. Queues are configured with a "heartbeat timeout"
  1. Clients: must send a "keep alive" if they are still wokring on the queue item
  1. Broker: on a restart, the "heartbeat timeout" will be rest to start counting at 00, but will eventually timeout if the client has also gone away
  1. Broker: use a default timeout of 5 minutes if nothing was specified by the comments
1. Each Enqueued message will recieve a unique ID
  1. Broker: will re-use IDs once they have been succesfully deleted or sent to the Dead Letter Queue
1. Connected client to the broker becomes disconnected
  1. All messages being processed by the client will be treated as Failed

Features for clients:
1. Each message for a queue is only processed by 1 client at a time
1. Each message must be responded to with sstatus
  1. ACK SUCCESS - item was processed succesfully and can be delted
  1. ACK FAIL - item failed to process and will be retried if it has not reached the retry count.
                If the retry count has been reached, it will be sent to the dead letter queue
1. Can subscribe to a unique name -> global, recieve messages from all tags
1. Can subscribe to a unique name + any number of tags:
  1. When new queues are created they will be applied to existing connections
  1. Tag Flags:
    1. Exact - subscribe to only the queue with the exact tags specified
    1. Strict - only queues that coontain all the tags will be considered
    1. Combined - subscribe to any queues that have any of the tags
    1. Any - subscribe to any queues that match any of the tags
