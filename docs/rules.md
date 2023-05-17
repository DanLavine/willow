Rules are a far off feature, but I think are a necessaty of a complex feature for this project. So it is
worth spending some time to make sure that these features can fit into the data structures I'm currently
creating.


0. General setup for rules
```
Groupings: (might not be needed. Just makes logical sense to view)
  1. Multiples rules that are all similar
    - groups // multiple Select clauses, all together
  2. An order is needed to know which rules take precident over others
    - order // this way we know which rules to act on. With the most restrictive rules being last

Select:
  1. Grouping of keys
    - group by - [key0, key1, key2, ...] // group any arbitrary tag groups by their keys

Where Clause:
  1. should be able to provide specific filters with [and, or]?
    - key value pairs // explicit key value tag to search for
    - key exists      // does this make sense? it should be in the where clause I think
    - exclusion       // ignore cases

Rule:
  1. Limits
    - max - max number of entries that can be using a rule
```
