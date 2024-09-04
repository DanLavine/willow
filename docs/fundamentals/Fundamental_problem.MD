Fundamental Problem
-------------------

The packages that are developed for Willow and the other related services, are built while trying to solve
a single problem: Managing highly available concurent operations around single locked operations.

    Lets take a look at couple places where this is easy to see:
    1. Locker - A distributed locking serivce, where all clients must eventually reconsile against a single
                source of truth for who has the single lock.
    2. Willow - When enquing data, if there are no clients yet processing any previous requests, then we want
                to update the old resource so they are ignored and save everything to disk. Also, any number
                of clients that eventualy come along and attempt to read the data will only obtain a single
                copy of what has been enqueued

    In both of these cases, we have a couple of common needs:
    1. A central point of writing to disk to save our state and survive system restarts.
    2. Ensuring that only 1 of N requests process at a time
    3. Persisting the request before procesing another


Which now brings up a general issues for API calls that can take a while to complete, especialy around DELETE
operations. In many cases if using a SQL or NoSQL DB and "deleting" a resource, that resource is some row (and
dependent relations) in the DB which is a quick operation. But also notrably, the DB blocks all other operations
against the same resources because it is locked.

    But, what if in the case of Willow we also want to clean all the disks we have enqueued data to when removing
    a queue? If we want the system to be HA and Horazontially scalable, then we need to account for N nodes where
    the delete operation will occur, and who knows how long it will take. That is now 100% dependent on the amount
    of data enqueued to the Disks.

    To try and address these issueswe can have some fundemental differences on how to complete these:
    1. When deleting, operations should use a "tombstone" which indicates that it is in the processes of being deleted
    2. APIs that notice a "tombstone" can quickly respond back to clients with the status of the operation in progress
       rather that having the DB be locked and eventually time out
    3. The main resource that has children is responsible for all the logic to handle child deletion
