Shared Lock
-----------

One larger feature I think that Locker service could currently make use of currently would be some
sort of "Shared Exclusive Lock".

## Current issue

Currently when `Willow` enforces the Rule limits for the number of enqueued items, all channels will all need to lock
the same `{"_willow_channel_name": "value"}` type of lock identifier. This can be very expensive operation if there is
a very large Queue with many channels.

## Solutions

1. Create persistent Locks with an initial "max number of concurrent lock operations"
  
  Features:
  * When creating a lock, it could be a persistent lock that exists even if no clients are waiting or have the lock
  * There can be an initial number of max clients that can all obtain the lock at once since they are guaranteed to pass
    the Limiter's Rules for max enqueued items.
  * If Willow succeeds in the process of enqueuing an item, then the lock's max concurrent connections can be decremented.
    This way we can keep track of the `Limit == max concurrent connections`. Similarly when an item is removed from `Willow`,
    this Lock would need to be incremented to allow for the same max number of concurrent connections

  Cons:
  * Feels like the Limiter logic is now spilling over into the Locker service.
  * We need to ensure these services are always in sync now.

2. Could a lock be upgraded from Read -> Exclusive, or decremented Exclusive -> Read would this allow for a different workflow?

3. Could Willow just be smarter and release the lock as soon as the "Max enqueued check" has passed?
  
  Origin:
  * Willow currently grabs all the locks it needs up front as I was scared a 3rd party could create the same Key + Value Pairs
    that conflict with the Rules setup by willow. But maybe that won't be a problem in the future when I have authorization and
    only services can use a `_` to begin Key + Value resources as a naming convention.
