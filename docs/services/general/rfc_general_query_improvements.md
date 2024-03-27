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

# Workflows to address the query issues

There are 2 main stratagies that I can easly think of to solve the issues below, but they each have there own
drawbacks and limitations in a horazontally scaled system.

1. Consistent hashing for `Key Values pairs` to know which node a resource belongs to

   **Pros**:
   1. Routing is somewhat easy to maintain as everything is consistent
   2. Only data that is save is at the source of truth on the Nodes
    

   **Cons**:
   1. If there are many "bad actor" Resources on a particular node, then they cannot be moved to a new isolated node.
   2. Scaling the system is hard as each resource meeds to be moved in some sort of automated fashion to the proper node
      and a thin routing layer has no knowledge if a resource has been moved or not.
      1. Possible solution: The "Nodes" would need to forward the request to the previous node in the consisten hash chain.
      2. Possible solution: The new node would need to wait to process a request untill a rebalance is comoplete
   3. Querys for `Order By` keys are very hard and slow as they need to reach out to each node and then perform an inital lookup for values
      currently exist. Then perform single actions on resource for specifically sorted `Key Value pairs`.
  
2. Thin API that defines the `Key Value pairs` the Resources and which Nodes they belong to. In this case, there will need to be a
   "DB" that has the objects defined and which nodes they run on. The "DB" still has the problems of solution 1, but are now much
   more managable as it is a thin data layer with no actual logic.

   **Pros**:
   1. Scaling the cluster can be easier as the "DB" layer could provide additional resource operations as
      a "deleting"/"scaling" etc state for the "thin" api to quickly respond in those scenarios
   2. Querys can be much faster as they need to now:
      1. Interact with the DB state which can prioritize fast execution
      2. Based on what is returrned, perform some sort of operation
    

   **Cons**:
   1. The data for a "Resource" is now duplicated. So when performing a "CRUD" operations for example we would need to:
      1. Lock down the "API DB" for the create operation
      2. Perform a create of the resource on the internal api (Possibly change the API a bit as it could just be an ID now?)
      3. Report to the DB that the resource is created
   2. Split brain scenarios seem much easier to come across now if something crashes before it is able to respond to the replication
      data that the "Create/Delete" operation successfully completed


# Unknowns

1. With current thoughts on authorization for `_[name]` keys, are those special cases? (I.E. when looking for a
   particular Rule to enforce from Willow, willow will have some sort of `_willow_queue_name` Key + Value. But
   what happens now when using the `KeyLimits` to enforce specific length of items we are searching for. Does the
   `_willow_queue_name` Key count against the Max Keys limit?) This will require some investigation on a pattern
   that can easily be described and makes sense
2. Another major use case I had in mind when designing Willow was a replacement for K8S's internal queue for processing
   deployments. Has the exact same feature set of processing only 1 deployment at a time, where each is defined by
   the "kind, namespace, label (name)". So that matches the current arbitrary key value tagging system, but what has yet
   to be accounted for is the Taints and Tolerations. That can be thought of "metadata" for the queue as key value pairs,
   but those are saved on the `channel` itself and can be updated without applying to the key values that define the `channel`
   1. Possibbly this does't make sense to do as fields directly as that could all be stored in the `ITEM's` data as the actual
      configuration to run that is pulled via a scheduler. Would need a way for infinite retries from Willow though.

# Work items

1. First thing I want to try, but have gone back and forth on before was adding an "Any" type to the EncapsulatedValues. This way
   the Limiter can be setup for the `GroupBy` to be similar Key + Value Pairs like all the other api calls and logic for
   the "Any" types can be the behavior for what `GroupBy` currently does. Now that the Apis allow for the behaviors I want, I think
   this can be cleanly done. Every time before was a bit confusing since I didn't have a clear idea of the actual service interactions
   1. In addition to this, I think the "Match" apis can go away and just become queries
   2. There is something nice about "Match" api behaviors though to explicitly see what the Rules/Overrides are for a particualr
      set of Key Values. Though that might be something more akin to a "dry run" api that explains the details behind the scenes
      thats going on.
