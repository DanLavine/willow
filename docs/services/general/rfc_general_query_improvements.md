General Query Improvements
--------------------------

# Problem

### Query Pagination and Order
Currently the common API models when querying can search for Key + Value combinations a user is interested in,
but this needs to be greatly improved to include additional functionality:

1. Imposing number of Resource to limits on a lookup on a query
2. Pagination so we don't need to grab the whole world in 1 go
3. Sort order by specific key
4. Round robin lookup
5. Random lookup

### Special Cases of API data manipulation
There are special cases to consider saving the data differently and or changing the API.

1. Limiter's Rules are defined by the `Name` and `GroupBy` fields. The `Name` defines the
   unique associated ID to use, and the `GroupBy` fieilds is a "special" tree where each value of the
   `GroupBy` is set to the empty string. Then all the `Match` querys take any of the "Keys" and match
   them to the empty strings in a first query to find the Rules, and a second query with the real
   values to find any Overrides. This feels dirty and can hopefully be addressed by what the service
   should really be doing for the exploritory APIs

# Workflows to maintain

1. Currently there is no "DB" like MySQL, PSQL, MonogoDB, etc as a data storage as each of these services have to much
   "in flight state" or all IO disk heavy that is the source of truth. So we need to ensure that any requests to a
   "resource" (I.E. lock, Willow queues, etc) always route to the same Node(s) cluster (failover or ha through something
   like RAFT). This is because a resource like a lock needs to manage all potental clients trying to obtain it at one central
   location, or a Queues Items to be enqueued are all written at the same location if they are updateable.
2. Any `Resource` is defined by the `Key Value pairs` which point to a single resource id that is expected to be
   persistent. In terms of APIs, if a resource is a relation (OneToMany or HasMany), thats defined through the urls
   such as `GET /v1/limiter/:limit_id/overrides/:override_id` (get Override)
3. If a `Resource` is expected to be automatically cleaned up and created on the fly, then the APIs can look similar
   to `POST v1/willow/:queue_name/channels` (enqueue item to a channel). In this case the channel is defined by the
   enqueued Item's `Key Value pairs` and can be created if there are no Items, or destroyed when the last item is removed.

# Workflows that this proposal attempts to answer

There are 2 main stratagies that I can easly think of to solve the issues below, but they each have there own
drawbacks and limitations.

a. Consistent hashing for `Key Values`
  1. With a consistent hashing of `Key Value pairs` then we would know which node a resource belongs to

  Pros:
    1. Routing is somewhat easy to maintain as everything is consistent
  Cons:
    1. If there are many "bad actor" Resources on a particular node, then they cannot be moved to a new isolated node.
    2. Scaling the system is hard as each resource meeds to be moved in some sort of automated fashion to the proper node
       and a thin routing layer has no knowledge if a resource has been moved or not.
        a. Possible solution: The "Nodes" would need to forward the request to the previous node in the consisten hash chain.
    3. Querys for `Order By` keys are very hard and slow as they need to reach out to each node and then perform an inital lookup for values
       currently exist. Then perform single actions on resource for specifically sorted `Key Value pairs`.

### Horizontaly scalable node selection for a resource

Lets take the simplest look at the resources, Locks. Each lock is defined by the `Key Values` that make up the lock.
What would be the proper


Unknowns:

1. With current thoughts on authorization for `_[name]` keys, are those special cases? (I.E. when looking for a
   particular Rule to enforce from Willow, willow will have some sort of `_willow_queue_name` Key + Value. But
   what happens now when using the `KeyLimits` to enforce specific length of items we are searching for. Does the
   `_willow_queue_name` Key count against the Max Keys limit?) This will require some investigation on a pattern
   that can easily be described and makes sense

# Solutions

Unknown for now. This will requires some experimentation to ensure:

1. Using sorts does not put the services into deadlock scenarios:
  * Since everything is an actual "processes" that is looked up and an operation is performed. We need to ensure
    there are no locks for the entire query operation to process? This is complex one to think through
2. the special case of `_[service]` Keys are accounted for