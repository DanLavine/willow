Limiter
-------

The `Limiter` process is a generic locking service for any group of Key Value pairs.

### API
TODO

### Logic
TODO

### Gurantees
TODO


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
	"Name": "rule2",
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


## Open Questions

1. How should the Create and Get operations define where items are saved in a horizontal scaled cluster 
```
    # POST /v1/limiter/rules
    Will create a rule, but in a HA world, the `rule_name` variable could dictate what node to land on.

    # GET /v1/limiter/rules/:rule_name
    This way any requests for APIs that contain the `:rule_name` can be routed to the proper server

    So In this case `rule_name` which is actually the _associated_id in the BtreeAssociated can be used as the "router" key to use since
    it is the one thing that is common for all requests.

    1. Do the API serversr just proxy the request to other servers?
    2. Do I have some sort of "proxy" level that understands how to pase requests for the _associated_id on create, and
       then on other opertions [PUT, GET, PATCH, DELETE], it knows how to read the URL for the key to route requets?
```

2. How should the "Have Many" relations be handled between Rules and Overrides in a horizontal scaled cluster
```
    # POST /v1/limiter/rules/:rule_name/overrides
    1. Will create an override, but in a HA world, the `:rule_name` variable should be first checked to ensure that the
       Rule actually exists. This would gurantee that a "ForeignKey" constriant holds true for child object operations.
       So this means all operations need to go through the server hosting the acual Rule

    # DELETE /v1/limiter/rules/:rule_name
    1. Will now delete the rule and we want to ensure "Cascading Delete" operations on all of the overrides.
        a. This is highly dependent on how "open question 1" was implemented. It could be all local delete operations
        b. Could also route to each other node in a fan out to trigger deletes for all Overrides, where they contain
           the rule's _associated_id (saved via the 'custom' data type). Would need a new "internal" api for this
```
