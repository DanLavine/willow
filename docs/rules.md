Limiter Rules
-------------

The limiter provides general record keeping of in progess work for any possible key value pairs. 

Feature Goals for the Limiter:
1. Be able to create a "general" rule that can be applied to any number of different tag groupings
2. the "general rule" should be able to query a set of tags to know if they apply to the rule
3. a rule can have a "tag grouping" overide where the key+value pairs match the "grouping"

Initail design to be acounted for:
1. There are a bunch of "rules" where the "name" of the rules provides as the btree key
2. Each rule has a "default" key limit, but the rule itself can be overriden
3. a rule can be provided an overide where the key+value pairs match the "grouping". Otherwise, all other groupings are assumed to be the default.
4. The "counter" for what is running is independent of the rule set as those are only for the limits. The values for what is actually running is a "combination" of all possible key+value pairs that make up a grouping
