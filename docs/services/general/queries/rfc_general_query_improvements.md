General Query Improvements
--------------------------

Current work in progress, investigating `Solution Considerations: 2`

# General Problem

There are currently 2 major query operations that I can se being useful, but I don't see an easy way combining them
into 1 general query model and operation. Lets define the main query operations by looking at how the same objects
can be queried with different behaviors

1. Relational Queries for the state of the system
   Example of this would be:
   ```
   We want to find all possible Channels for a Queue on the Willow service
   ```

   In these cases, we are potentialy obtaining a large data set and need the common pagination tools:
   ORDER BY ... ASC/DESC, LIMIT.


2. Actionable queries against an object in the system
   Example of this would be:
   ```
   We want to dequeue an Item from a Channel on the Willow service
   ```

   In this case we are working against the same data set as the last, but now the the pagination tools are not
   paricularly helpful. In the case of attempting to dequeue an item, with LIMIT=5, we would connect 5 random
   channels, but if they have all hit the `Limiter's` resource constraints, we won't actually dequeue anything.
   Also, Willow would need to be smart enough to try and find a new channel if one was deleted. Having those common
   pagination fields means a lot of custom logic per service on what those fields mean. It will also resort in many
   more API calls as a `SELECT from QUEUES WHERE "ORG" = "abc"` is a general query that can be sent to each HA node,
   and the first to respond will be selected. But `SELECT from QUEUES WHERE "ORG" = "abc" ORDER BY "ID"` would mean
   we need to query all possible value and make a request in order for each possible operation. It is doable, but
   I don't think there would be a nice way to manage this in large data sets and is overly complicated

3. Pagination for single KeyValue resources
   Example of this would be:
   ```
   We want to list all the queues in Willow that are defined simply by the `Name` resource
   ```

   If there are many queues, we still want to query for `SELECT * WHERE name >= [last id] && ORDER BY 'name' ASC && LIMIT 5`.
   But there is no notion for the end user that `name` is the field as they did not define it. Perhaps the APIs need to change
   to always allow a collection of Key + Value pairs and then the 'IDS' are always used for the API operations?

4. Need a way of also declaring the last `AssociatedID`

   When attempting to list all the items of something in the DB, if they don't have a `ORDER BY` key, then the
   `AssociatedID` will be used. So we need a way of slecing the last value for that as well... Perhaps that the only thing
   that needs to be provided when performing pagination?

# DB Rules

1. We record the `_associated_id` for all items saved in the trees and just return that in the future?
2. For recording child objects, we can allways have the `_[resource]_associated_id` be used to know the parent object's value on a one-to-many relation
3. These associated `_[resource]_...` fields can be added on the fly for many-to-one relations?

# Query Rules

1. the specific fields of `_` wouldn't need type restrictions as they are always strings, but maybe they are still required for a nice api?
1. When declaring Min/MaxNumberOfKeys these do not account for any internal fields (begin with an `_`)

# Workflow Problems for specific services

### Willow

1. Dequeue operations

   There is a need to handle possible dequeue selection, such as:
   1. Round robin
   2. Random
   3. Longest time since last ran
   4. Priority

   I can see a use case or need for all these possibilites. Also some are more client driven (Priority, Time since last ran)
   and others are somewhat service driven (Round robin, Random, Time since last ran)? The Longest time since last ran, could
   be some sort of sort order by a `last updated` key or something that is common on the DB, or just a service configuration?

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

# Solution Considerations

1. With the current APIs already, could we split them into a "watch" vs a "get" api where `watches` are slow and don't allow
   for the pagination featues, where as `get` return quickly and support all the pagination? This way our clients when finding
   an item to dequeue can be explicit about what they are finding... This is willow only, not thinking about the limiter /w its
   'match' actions. This also needs to not use the `LIMIT`, that would be awful.

2. So if we don't have any pagination options, could the client's queries perform a subset of operations per node:
   SORT BY, ORDER ASC. This means we don't 100% gurantee to get the highest priority, but we could get the highest priority
   in the node when an action occurs. So all that being said, I think the "Action" vs "Relational" apis means `SORT BY` and
   `ORDER ASC` have slightly different contexts because 1 is guranteed. Should we use differnt words for the API models to
   ensure that the user knows something different is occuring?