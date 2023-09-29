Locker
------

The `Locker` process is a generic locking service for any group of Key Value pairs.

## How Locker Workes
TODO

### API
TODO

### Logic
TODO

### Gurantees
TODO

## Open Questions

1. On a server restart, what should happen to clients waiting for locks that are partialy thrrough a request?
    a. Should they be dropped and cleared from disk?
    b. Do we only wite to disk when we "have" all locks -> then respond -> shutdown
    c. Can we "clean up" partially written data, that was never stored?

2. Clients will need to re-establish which locks they actually have acces to using an "identifier"
    a. "identifier" -> datatypes.StringMap


## Usage

This service is being built out as the logicial share lock holder for the `Limiter`. Each action in the `Limiter` needs
to grab all possible Key Value lock pairs on any Rules check operation. This was overlapping rules with any number of
different key values enforce that there is no race condition when checking any particualr Key Value pair.

Example:

Rule 1:
```
{
	"Name": "rule1",
	"GroupBy": []{
        "key1", "key2",
    },
	"Seletion": {},
	"Limit": 3
}
```

Rule 2:
```
{
	"Name": "rule1",
	"GroupBy": []{
        "other key",
    },
	"Seletion": {},
	"Limit": 1
}
```

Then assume any number of parallel reuests such as:
Rule check 1:
```
{
    "KeyValues": { 
        "key1": {"DataType": 12, "Value": "string1"},
        "key2": {"DataType": 12, "Value": "string2"},
        "other key": {"DataType": 12, "Value": "string3"}
    }
}
```

Rule check 2:
```
{
    "KeyValues": { 
        "key1": {"DataType": 12, "Value": "string1"},
        "key2": {"DataType": 12, "Value": "string2"},
    }
}
```

Rule check 3:
```
{
    "KeyValues": { 
        "other key": {"DataType": 12, "Value": "string3"}
    }
}
```

1. In This case, we wouldn't want `Rule check 1` to pass the `Rule 2` limit, while then also checking `Rule 1`.
   If that was to pass concurrently while another `Rule check 3` was running in parallel and succeeded, then 1 of the 2 rules
   must be in violation of `Rule 2`. This means that we need to ensure that all rules grab all possible lock combinations
   as then ensure that rules are enforced