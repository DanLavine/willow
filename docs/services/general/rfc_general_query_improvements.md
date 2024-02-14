General Query Improvements
--------------------------

# Problem

Currently the common API models when querying can search for Key + Value combinations a user is interested in,
but this needs to be greatly improved to include additional functionality like:

1. Pagination so we don't need to grab the whole world in 1 go
2. Sort order can allow for Priority

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