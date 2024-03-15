General Query notes
-------------------

# Sorting /w pagination

When thinking about sorting, I think that we need to have some general rules for how we can possibly sort values:
1. KeyValues of saved items in the tree need to be hashed consistently
    1. Items with more KeyValues are hashed to a longer value
2. Sorting on a particular "key"
    1. Items with less KeyValue pairs will be sorted by first (by default)
3. Should there be a general "length" sort option for AssociatedKeyValues?