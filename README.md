# willow

Willow is a message broker that can be used as either a pub-sub or message queue system.

# Terms
TODO - not 100% sure on what I want to call the specific pieces of this system yet.

* Broker - A messaging system that can be broken down into the following
  * Queue   - a first in, first out queue the sends a message to one Consumer
  * Pub-Sub - a general pub-sub message bus that distribues messages to all Consumers
* Message - A unit of work that is either of the queue or pub-sub brokers
* Producer - This is the client that sends messages to willow
* Consumer - This is the client that receives messages from willow


# Unique Features

What makes willow unique from other systems and why should I use it rather than something
like Kafka, RabbitMQ or NSQ?

### Queue

Queues can be configured with an updatable message configuration. This means that if there is
currently a message in a queue, that has yet to be processed. The incoming message will replace
the message currently in the queue.

This can be used for a number of use cases like:
1. CICD only wants to build the "latest" commit, but each incoming change can easily be published
   to the queue.
1. Long running update operations that have multiple changes stacked can all be collapsed into the
   latest update operation, skipping any middle operations that are no longer needed or valid.

# Considerations

1. Should this use something like amqp for the connection protocol?


# Work List

## In Progress
1. can ack a queue item

## TODO (in order)
1. when a queue item fails, it goes to the the dead-letter-queue
1. add a retry ablity for queue items
1. add a "timeout" ability for queue items
1. dead-letter-queue is "updateable"
  1. any item currently being processed and waiting for an ACK() is not updateable
1. dead-letter-queue can load from disk
1. Add ability for queues to have limits (number of messages)
1. ensure metrics are working properly
1. Split out to a broker/node infrastrcuture as 1 part of HA
