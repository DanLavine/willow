# willow

Willow is a message broker that can be used as either a pub-sub or message queue system.

# Terms
TODO - not 100% sure on what I want to call the specific pieces of this system yet.

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
