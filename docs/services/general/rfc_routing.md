Routing in HA
-------------

# Problem

How do I address the routing issue for HA when determining where to a request should go?

# Address the routing issues

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
  
2. Thin API that defines the `Key Value pairs` the Resources and which Nodes they belong to. In this case, there will
   need to be a "DB" that has the objects defined and which nodes they run on. The "DB" still has the problems of
   solution 1, but are now much more managable as it is a thin data layer with no actual logic.

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
