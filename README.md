# willow

Willow is a message broker that can be used as an updatable queue (and someday eventually pub-sub message system).

# Terms
* Broker - A messaging system that can be broken down into the following
  * Queue   - a first in, first out queue the sends a message to one Consumer
    * TagGroup - Each queue is created via number of "tags" that define specific item details.
* Item - A unit of work that is sent or pulled from the queue
* Producer - This is the client that sends messages to willow
  * TagGroup - when sending Messages, all tags must mutch exactly 1 queue that will recieve the message
* Consumer - This is the client that receives messages from willow
  * MatchRestrctions - when consuming message, there are a number of stratagies. See [#consumer_setup]


# Unique Features

What makes willow unique from other systems and why should I use it rather than something
like Kafka, RabbitMQ or NSQ?

## Updatabel messages

If a message has been published to a queue it can be configured to be "updatable". As long as no
consumers have processed this message, then the next message sent to the queue can overwrite the
unprocessed message. This way any clients that eventually pull from the queue just retrieve the
latest message that needs processing.

Use Cases:
1. CICD only wants to build the "latest" commit, but each incoming change can easily be published
   to the queue without having to worry what will run.
1. Long running update operations that have multiple changes stacked can all be collapsed into the
   latest update operation, skipping any middle operations that are no longer needed or valid.

## Consumer Setup

Consumer can easily be setup to pull from a number of queues (even if they don't yet exist).
When crating the consumer, we specify what tags we are interesed in, as well as the consume stratagy:
1. STRICT - only pull from the exact queue that matches all the tags
1. SUBSET - pull from any queue that that includes all provided tags
1. ANY    - pull from any queue that contains any provided tags
1. ALL    - pull from all queues

Use Cases:
1. Initial thought would be for releasees where you want to build executables on certain OSes
   (Mac, Windows, Linux), which can all be seperate tags on queues. Then each of those queues might
   have different features that need to be tested (I.E, we are releasing a video game and want to test
   different graphics cards, cpu arch, etc). We could have specific queues that specify more in grain details
   about all the combos we want to test. From here our Consumers could be quite varied as the server rigs
   have multiple setups all configured to test various hardware. We halso ave a wide spread of potential releases
   from any CI pipeline for N repos + N branches. Pulling from a SUBSET for our server machines configuration
   can just run any build from any queue