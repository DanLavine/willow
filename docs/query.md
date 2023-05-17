Need to figure out what I actully need for a query. I think there are 2 places where a generic query
makes sense, but I'm not sure on if both can use the same data structures, so thats what this is going
to try and figure out.

0. General rules for query
```
Select:
  1. Types
    - Strict // all tags must match. needs to be key + value pairs
    - Subset // all tags must be included
    - Any    // any of the tags requested can be found
    - All    // get all possible tag groups

Where clause
  1. should be able to group multiple with [and, or]
    - key value pairs             // explicit tag search 'name = dan'
    - key exists                  // check to see if a tag exists 'name != nil'
    - exclusion (previous rules)  // ignore all cases 'name = dan'

Limits
  1. tag restrictions // I think these hurt and are not needed
    - max number of tags [inclusive] // max number of key value pairs we want to select from
    - min number of tags [inclusive] // min number of key value pairs we want to select from
```

