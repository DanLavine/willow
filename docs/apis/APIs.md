APIs
----

Please refer to the [fundemental problem](./../fundamentals/Fundamental_problem.MD) to familurize yourself with the
initial problem this api pattern is attempting to solve.


# Questions to solve

1. KeyValues for Definitions (DBDefinitions)

~~    For certain `Keys` there are times where we want those keys to be unique, like the in the `Willow` API's for queues,
    those should have a "name" key that is unique for all queues (so unique indexed like SQL).

    Are there times that we want to have just specific values to be unique?
        - need to figure this out
        - this is the shared "lock" for the locker, so maybe its just the single key and the litem is a lock?
        - could this be enforce on "willow" itself? that means that we can never change the "keys" though, so
          it doesn't seem great.~~

~~2. Can the DB's take a "schema" to enforce like SQL/NoSQL DB migrations?

    This could be a way to solve API validation enforcement. It would also allow for "migrations" when wanting
    to blanket change certain Key + Value pairs.

    This would be nice, but I don't think its important to do for now~~

This is imposible to do. This is because in an N node configuration, we need to know the Key + Values url maping to the node.
If there are N nodes and 1 of N keys needs to be unique, there is no way to enforce that across all the nodes. It needs to
be enforce across the entire Key + Value pairs.

So we can just do an API enforcment for specific Key + Values if we want to have a guranteed Key + Value pair to exist