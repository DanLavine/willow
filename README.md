# Willow

TODO: add a content table to jump to sections

Willow's aims to be a different message broker that provides a few extra features not commonly found
in many of the modern brokers one normally thinks of such as `Kafka`, `RabbitMQ`, or `Redis message broker`.

Such as:
1. Message waiting in the queue can be updated if there are no clients yet processing the message
1. Generic tagging for any message in a queue 
1. Generic limits for concurrent processing to be placed on any number of messages through the tagging system


To do this, there are 3 Main components:
* Willow - Message broker that can be used to generate queues with any unique tags for those queue
* Limiter - Enforces rules for any tag combinations that limits how many tag groups can be running at once
* Locker - Service that the Limiter relies on to ensure that competing resources for the same tag do not race each other

Let me try to explain with an example workflow I wish I had many times working with a CICD system and where these pieces
can fit into this.

#### Example Use Case

When working on any new Mobile App for a small team which has grown to 10 developers, we set up a CI cd
system that has been able to easily pull all the branches from Git and run the unit tests for the product
no problem. 

But it is at this point and scale where we start setting up a few real world integration tests, to ensure
the products stability. So perhaps in this case, we want to ensure that we have proper hardware:
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
3. Stop running so many integration tests and be less stable (This sadly is the one we see most common :( )
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
product and do not scale out. One example where this is hugely present is with the Kubernetes Service.

Anyone who has ever deployed a `Deployment`, `ReplicaSet`, `DaemonSet` have take advantage of the 3 features:
```
1. Message waiting in the queue can be updated if there are no clients yet processing the message
2. Generic tagging for any message in a queue 
3. Generic limits for concurrent processing to be placed on any number of messages throught the tagging system.
```

* feature 1 - is solved where you cannot deploy the same `Deployment`, `ReplicaSet`, `DaemonSet` if one is already processing
* feature 2 - All of the tags are in [ `k8s namespace`, `k8s label.app`, `k8s kind (which is Deployment, ReplicaSet, etc.)`].
              It is also important to note that each of these all generate a `unique` item and are case sensitive.
* feature 3 - In the world of K8s, the Limit for each of these operations is 1 and is not configurable

## So how can Willow, Limiter, and Locker enable these workflows

### Willow

### Limiter

### Locker


## Building and Running

See the `docker` directory

## APIs

All the service have a /docs endpoints. need to figure out how to display the `docs/openapi` nicely in github.