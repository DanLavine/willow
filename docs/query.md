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


When selecting the readers (inclusive query)
```
Select * FROM deployments // global
Select * FROM deployments WHERE namespace == default // select everything in the default namespace
Select * FROM deployments WHERE namespace == default LIMITED BY KeyValuePairsMin = 2 KeyValuePirsMax = 2 // select everything in the default namespace when there are only 2 tags

// This is super hard to implement... Thats because we could get unique tags that don't yet exist so we don't have the channels
// This makes sense on the rules, but not for the readers? Also would be super slow for each request which isn't good
Select * FROM deployments WHERE namespace != default // select everything not in default namespace

// maybe its better to fx the readers for clients setup.
```

When creating the rules
1. How do rules get applied to queues that already exist?
2. Should rules be applied accross queues?
   1. I.E. coud have 1 windows, 1 mac and 1 linux build queue, but any combo of running builds for a team should only == max 5
```
SELECT * FROM deployments // global rule that evey group needs to adhere too
SELECT * FROM deployments WHERE namespace == default AND deployment != prod // eveything other than prod
SELECT * FROM deployments WHERE namespace == default AND KEYS name EXIST
```


So I think thats what I need as a query. A new key word that makes sense for 3 differnt ways of uerying:
1. Explicit selection (I.E. WHERE namespace == default)
2. Checking if a key exists (I.E. WHERE KEYS namespace EXIST // this is a generalized Query where it can include items that don't yet exist in a DB
3. Negating the previous 2 selections (I.E. [WHERE namespace != default | WHERE KEYS namespace DONT EXIST] ) // this is like normal Relational Queries that find all items in a DB

NOTE: haveing 2 + 3 together isn't good (for my readers). Maybe thats the real issue. how do I solve that problem. Since
2 is really (WHERE namespace != nil) in theory....
