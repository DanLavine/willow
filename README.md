# Willow
[godoc](https://pkg.go.dev/github.com/DanLavine/willow)


Willow is a message queue that aims to provide a rich feature set for slower workloads. In the current development
lifecycle this is currently a POC and drives out the intended features.

Main features:
1. Queue's enqueued Items waiting to be processed can be updated if there are no clients currently working on them
1. Items are defined by a collection of generic Key + Value pairs
1. Limits for concurrent processing restrictions can be placed on any combination of the Key + Value pairs
1. Clients can query for any Items in a Queue that match particular Key + Value pairs

To do this, there are 3 components:
* Willow - Queue that process any Items and checks the Limiter to know when an item can be processed
* Limiter - Enforces rules for any Key + Value Pair combinations that can be running at once
* Locker - Service that the Limiter relies on to ensure that competing resources for the same Key + Values do not race each other


# Table of Contents

- [Quick Services overview](#quick-services-overview)
  * [Locker](#locker)
  * [Limiter](#limiter)
  * [Willow](#willow)
- [Building and Running](#building-and-running)
- [Limitations](#limitations)
- [Future Plans](#future-plans)
- [TODO](#todo)

# Quick Services overview

The first service listed `Locker` is the simplest service and does not have any dependencies so its the easiest to explain.
From there each of the other services depends on the previous service to operate.

### Locker
[api](https://danlavine.github.io/willow/docs/openapi/locker/)

[full documentation](./docs/services/locker)

The Locker Service that provides distributed locks for any other services that need to manage competing resources.
I don't believe this would be reachable for any clients outside of a normal deployment for Willow as it serves as the internal
locking mechanism for the `Limiter`. This could change going forward as locks serve a more complicated feature set other than
simply Lock/Unlock.

### Limiter
[api](https://danlavine.github.io/willow/docs/openapi/limiter/)

[full documentation](./docs/services/limiter)


Limiter provides a way of creating generic runtime enforcement polices for collections of Key + Values. The Limiter service
requires the `Locker` service to be up and running for shared distributed locks. `Locker` ensures that different
key values being compared in parallel don't conflict with each other

### Willow
[api](https://danlavine.github.io/willow/docs/openapi/willow/)

[full documentation](./docs/services/willow)

The main drive for Willow was to develop an Item queue where any enqueued Items can be 'updated' in place as long as
no Consumers have picked them up for processing. While this feature alone might have been good to add to any other current
queue service that already exists, I built out a query api for dequeuing Items that have any number of Key + Value Pairs a
consumer might want to use. This allows for a flexible queuing service and Consumer clients can be very specific on what
they can operate on when paired with the Limiter


# Building and Running

See the `docker` directory for the full instructions to build and run all of the Willow services. Currently I am not publishing
any of the images to dockerhub since there are to many changes going on till I start tagging releases and they are reasonably
stable. At that pooint I will also want to setup a CICD to build and deploy things properly than doing everything manually for now.

# Limitations

1. Currently the system is only in memory only. I had to figure out how to save all my apis through the unique Key + Value Pairs
2. No authorization for Limitations created in the Limiter Service. So removing arbitrary things will probably cause major
   problems

# Future Plans

Currently I see this project being a nice dev tool and would like a single repeatable service to be open source to drive
out features that people might want. That being said, having a horizontally scalable solution and highly available fail over
solution to be the real selling point of this service. So for now, that documentation and work will remain private. If I cannot
get that funded this might just stay my constant passion project and can bring that over in time


# TODO

TODO is a list of work items that I plan on working on in a somewhat expected order. Now that `Willow` is operational
in an end to end system I want to keep everything running smoothly. For full documentation on each of the services and where
I would like to take them, you can check out the docs for each service directly. This small section is a highlight for what
I want to work on at a glance:

1. ~~Have the APIs use an `Any` data type that can compare against the Key + Value pairs. This allows for the Limiter to be a queryable api~~ DONE
2. Determine if there is a clean way to have all the `Query` APIs use common query feature set (ORDER BY, LIMIT, etc)
   There is greater documentation [here](./docs/services/general/queries/rfc_general_query_improvements.md) that highlights the major
   issues and their resolutions, or at least the work in progess.
