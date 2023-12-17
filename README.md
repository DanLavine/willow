# Willow

- [Featres](#features)
- [Example Use Case](#example-use-case)
- [Other use cases that already exist](#other-use-cases-that-already-exist)
- [So how can Willow, Limiter, and Locker enable these workflows](#so-how-can-willow-limiter-and-locker-enable-these-workflows)
  * [Willow Service](#willow-service)
    * [Willow publishing client](#willow-publishing-client)
    * [Willow subscribing client](#willow-subscribing-client)
  * [Limiter Service](#limiter-service)
    * [Limiter Rules](#limiter-rules)
    * [Limiter Counters](#limiter-counters)
  * [Locker Service](#locker-service)
- [Building and Running](#building-and-running)

# Features

Willow aims to be a different message broker that provides a few extra features not commonly found
in many of the modern brokers one normally thinks of such as `Kafka`, `RabbitMQ`, or `Redis message broker`.

Main features:
1. Message waiting in the queue can be updated if there are no clients processing the message
1. Generic tagging for any message in a queue 
1. Generic limits for concurrent processing to be placed on any number of messages through the tagging system


To do this, there are 3 components:
* Willow - Message broker that can be used to generate queues with any unique tags for those queue
* Limiter - Enforces rules for any tag combinations that limits how many tag groups can be running at once
* Locker - Service that the Limiter relies on to ensure that competing resources for the same tags do not race each other

Let me try to explain with an example workflow I wish I had many times while working with a CICD system. That
coul be improved with Willow's features

#### Example Use Case

When working on any new Mobile App for a small team which has grown to 10 developers, we set up a CICD
system that has been able to easily pull all the branches from Git and run the unit tests for the product
no problem. 

But it is at this point and scale where we start setting up a few real world integration tests, to ensure
the product's stability. So perhaps in this case, we want to ensure that we have proper hardware:
```
1. IOS device - attached to a Mac laptop to run
1. Android device - attached to a Linux server to run
1. Server - runs on a Linux machine
``` 

Even after setting up the initial hardware to start running our integration tests, the CI system has a slight
problem, but maybe it is still manageable at this point. That is there is only 1 IOS/Android device we have setup,
so only 1 instance of the `Main` branch with all our integration can be running at once.

There is now commonly a few ways the system is currently solved to handle this limitation, especially
as the team grows:
```
1. Start adding more hardware when CI becomes far to slow and commits start pilling up
2. Become more strict as a company as to what is promoted to the `Main` branch making for larger commits
3. Stop running so many integration tests and be less stable
```

But now what if there was a 4th option which was to allow for any items waiting in the `Main` branch's CICD
queue to be squashed so only the latest commit runs. This way no matter how large the team grows at least the
workflow for CICD is to run the latest commit each time, knowing that everything at the top of the commit branch
is working or not. `Willow` can solve the first workflow case:
```
1. Message waiting in the queue can be updated if there are no clients yet processing the message
```

**Example Continues!**

Now our team has gone through an additional round of funding and we have grown to 100 devs! Our CICD system is
now going to hit another problem. Do we really need to run every single commit across every single branch? Especially
when a lot of commits are just 'works in progress' as developers try to figure things out that are most likely unknown
or broken to them.

In this case, I want to use the other features of `Willow's` release, the `Limiter` service. Being able to limit the
combination of unique tags like: [`branch name`, `repo`] (these 2 details together define a branch to run in CI), we
could set a limit to 1 for each of those combinations. This way that if any branch is constantly committed to. It once
again starts to behave the same way as the `Main` branch, just run the latest commit sha. Now solving the last 2 features
```
2. Generic tagging for any message in a queue 
3. Generic limits for concurrent processing to be placed on any number of messages throught the tagging system.
```

#### Other use cases that already exist
Of course there are some use cases where these features are already present, but built specifically for the
product and do not scale out. One example where this is present is with Kubernetes.

Anyone who has ever deployed a `Deployment`, `ReplicaSet`, `DaemonSet` have take advantage of the 3 features:
```
1. Message waiting in the queue can be updated if there are no clients yet processing the message
2. Generic tagging for any message in a queue 
3. Generic limits for concurrent processing to be placed on any number of messages throught the tagging system.
```

* feature 1 - is present where you cannot deploy the same `Deployment`, `ReplicaSet`, `DaemonSet` if one is already processing.
              As a user, you can still attempt to deploy any of these as much as you would like. But its only once the current
              operation is done processing that the next operation happens. Which is the last configuration published
* feature 2 - All of the tags are in [ `k8s namespace`, `k8s label.app`, `k8s kind (which is Deployment, ReplicaSet, etc.)`].
              It is also important to note that each of these all generate a `unique` item and are case sensitive.
* feature 3 - In the world of K8s, the limit for each of these operations is 1 and is not configurable

## So how can Willow, Limiter, and Locker enable these workflows

### Willow Service
Willow is the message Broker service and can be used to coordinate any number of client to possible queues that match their tags

For a complete API list see the OpenAPI doc here // nothing atm till it is finalized

#### Willow publishing client
First to create a Queue in Willow, we need to define a Queue with a unique name:
```
# TODO: once api is finalized
```

From here, we can enqueue any messages onto the queue which are defined by their provided tags:
```
# TODO: example of message request with unique tags.
# TODO: comment on the 'updatable' tag to know if a message is updatable
```
In this example, we enqueued an item for our example CICD system. The tags [`repo_name`, `branch_name`] can be used to infer where
the origin of the queue came from and part of the scheduling. The other details [`os`, `ios_required`, `android_required`] can 
also be used as part of the scheduling and provide a more detailed view for what is actually needed to run the required tests.

If we were to make the same request again, but slightly different data (commit sha to run):
```
# TODO: once api is finalized
```
Willow will collapse the previous message if it has not yet been picked up by a subscribing client

If we wanted to ensure that a message was processed, we could set the `Collapsible` parameter in the request to false (I.E:
done in a web UI where a user really want to run a particular build ranther than the automated enqueue messge for every commit).
This will still collapse the last message if possible, but on the next message to come in for this particular set of tags,
the messages will be enqueued, so there are 2 total items to process.

#### Willow subscribing client
For the subscribing client, we can make a request to query all possible Tags that a queue has for something to run:

Example 1:
```
# This API should be used to query for any queues, where there are no IOS or Android devices
# TODO: once api is finalized
```

Example 2:
```
# This API should be used to query for any queues, where there are IOS devices, but no Android devices
# TODO: once api is finalized
```

Example 3:
```
# This API should be used to query for any queues, where there are are Android devices, but no IOS devices
# TODO: once api is finalized
```

Example 4:
```
# This API should be used to query for any queues, where there are are Android and IOS devices
# TODO: once api is finalized
```

Each of these client queries can be coming from any CI machines that describe their features. So Example 2 could be a machine
which has an IOS device attached and can run any builds that have those requirements based off the arbitrary tag [`ios_required`]

It is important to note that the Willow Service is smart enough to know that if a particular Queue doesn't yet exist, but in the
future matches a client's requested tags. Then the client will be able to receive the new queue's message. Also, before each
messag is dequeued, `Willow` checks the `Limiter` service to ensure no rules have met thier limits

### Limiter Service
Limiter provides a way of creating arbitrary rules for groups of `tags`. The Limiter service
requires the `Locker` service to be up and running for shared distributed locks. `Locker` ensures that different
Queues + Tags from don't conflict with each other.

The description below is a simplified API as it doesn't include any query operations and only provides an example
of the feature set one would want to use the `Limiter`service for.

Full api documentation can be found [here](https://danlavine.github.io/willow/docs/openapi/limiter/)

#### Limiter Rules
Each `Rule` in the Limiter service defines how to group arbitrary tags together. So in the case of our CICD
system where we only want 1 build running for each branch we could setup:
```
POST /v1/limiter/rules -d ' {
  "Name": "limit_cicd_branch_builds",
  "GroupBy": ["repo_name", "branch_name"],
  "Limit": 1
}'
```

Now, any collection of tags that share the same [`repo_name`, `branch_name`] all point to the same `Rule`
set. For example, a collection of `Counters` (described bellow) could be:
```
# all these when trying to run at the same time, would check the same branch rule, and only 1 can succeed
'{"repo_name":"willow", "branch_name":"docs", "os": "linux"}' 
'{"repo_name":"willow", "branch_name":"docs", "os": "windows", "android_required": true}' 
'{"repo_name":"willow", "branch_name":"docs", "os": "mac", "ios_required": true}' 

# This is different because the unique pair is named something different
'{"repo_name":"willow", "branch_name":"awesome-feature"}' 
```

To Account for the `'{"repo_name":"willow", "branch_name":"main"}'` though which maybe we want to run any number of instances
we can set an `Override` for the previous rule:
```
POST /v1/limiter/rules/limit_cicd_branch_builds/overrides -d ' {
  "Name": "increase_main_branc_limit",
  "KeyVales": {
    "repo_name":"willow",
    "branch_name":"main,
  },
  "Limit": 9001
}'
```


Now, we increased the number of parallel builds specifically for the main branch up to 9001! But that probably wouldn't
ever hit in reality as the number of clients to `Willow`` would need to support that much parallel capacity. Speaking of,
what about the limit of IOS and Android device? Well, we can just make more rules:
```
POST /v1/limiter/rules -d ' {
  "Name": "limit_ios_concurrent_bilds",
  "GroupBy": ["ios_required"],
  "Limit": 4 // whatever the physical hardware limit would be.
}'
```

Now, any builds that want an IOS device will be limited as well by the Willow service:
```
# all these now conflict on the new IOS rule
'{"repo_name":"willow", "branch_name":"experimental-1.3.4", "ios_required": true}' 
'{"repo_name":"willow", "branch_name":"main", "ios_required": true}' 
'{"repo_name":"willow", "branch_name":"stable-1.0.0", "ios_required": true}' 

```

#### Limiter Counters
`Counters` are used to check agains an `Rules` that might exist and their overrides. To increment a counter we can use:
```
POST /v1/limiter/counters -d ' {
  "KeyValues": {
    "repo_name":"willow",
    "branch_name":"main",
  }
}'
```

Each time a counter is incremented, any `Rule.GroubBy` is queried to see which `Rules` a particular `Counter` might match
against. Then, when matching against each `Rule`, the `Overrides.KeyValues` are queried to see which `Overrides` match against
the `Counter`. If everything is below the limit, then the `Counter` is either created or incremented by 1.

Code wise, there are a few optimizations that are made here. For example, if there is a `Rule` or `Override` that sets an explicit
limit to 0, then we know that we can stop processing the request and return a Limit reached error because nothing will process.
It is up to the client to eventually retry the increment, ideally with an exponential back off.


Lastly, `Counters` can be decremented though:
```
DELETE /v1/limiter/counters -d ' {
  "KeyValues": {
    "repo_name":"willow",
    "branch_name":"main",
  }
}'
```
 If a counter was to be set to 0. It is removed from the Limiter service entierly.


### Locker Service
The locker service is a simple distributed locking service that for now, I don't believe this will be reachable
from any clients as it serves as the internal distributed locks for the `Limiter`. This could change going
forward as locks serve a more complicated feature set other than simply Lock/Unlock.

Full api documentation can be found [here](https://danlavine.github.io/willow/docs/openapi/locker/)

## Building and Running

See the `docker` directory