Limiter
-------
[api](https://danlavine.github.io/willow/docs/openapi/limiter/)


# Terms
<dl>
<dt>Key</dt>
    <dd>Is always defined by a string</dd>
<dt>Value</dt>
    <dd>Is an object that defines what the type is (I.E. int, int32, float64, string, etc) and the data for that type</dd>
<dt>Key + Value Pairs</dt>
    <dd>a collection of Key + Values</dd>
<dt>Counter</dt>
    <dd>structure with a a collection of Key + Value Pairs and a number indicating the addition or subtraction of the total count associated with the Key + Value Pairs</dd>
<dt>Rule</dt>
    <dd>Restrictive resource that defines the the max limit for how many Counters can be created at once for generalized Key + Value Pairs</dd>
<dt>Override</dt>
    <dd>A specific specification for a Rule that defines an exact Key + Value Pairs limit</dd>
</dl>

# Service Features

Limiter provides a generic restriction enforcement service for any grouping of **Key + Value Pairs**. It does this by allowing
**Rules** to define which **Keys** when grouped together define a specific policy to enforce. Then any **Counters** who's
**Key + Value Pairs** mach the **Rule** or **Overrides** are matched against all other **Counters** who provide limits.

To try and explain this workflow, I think it is easiest to go through an example of how the Limiter Service can be used

## Example Usage

1. Setup for the use case

	Consider we are setting up a CICD system for this `Willow` Service hosted on github. If we wanted some sort of general runtime
	rules to limit the running resource because we only have 4 runners. Also, we want to ensure that the service builds on an instance
	of the latest Major OS instances (1 Windows, 1 Mac, 2 Linux).

2. Limit for how many build per branch can be running at once

	Assuming we want to utilize our resources to try and cover as many branches as possible while also ensuring that the Main
	branch can always build, we can create a **Rule** so that any branch can only be running only one build at a time
	
	```Rule
	# POST /v1/limiter/rules
	{
		"Spec": {
			"DBDefinition": {
				"Name": "limit branch builds",
				"GroupByKeyValues": {
					"repo_name": { "Type": 1024 }, // 1024 means 'Any'
					"branch_name": { "Type": 1024}
				}
			},
			"Properties": {
				"Limit": 1
			}
		}
	}
	```
	
	Now in this case, when attempting to pull a commit from any of our  4 Runners, we can try and setup a **Counter**:
	```
	# PUT /v1/limiter/counters
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"}, // type 13 is a string
					"branch_name": {"Type": 13, "Data": "new_feature_72"},
					"os": {"Type": 13, "Data": "windows"}
				}
			},
			"Properties": {
				"Counters": 1
			}
		}
	}
	```
	
	So in this case, when creating the **Counter**, the Limiter Service enforces 
	```
	{"repo_name": "willow", "branch_name": "new_feaure_72"}
	```
	that no other resources are running at the same time. Since as there are no other **Counters**, this would return OK.
	If however there was another item already processing for this branch, then the CICD runner would attempt to create a
	new **Counter** and that would fail as `{"repo_name": "willow", "branch_name": "new_feaure_72"}` is over the **Rules** defined 	
	```
	"Limit": 1
	``` 
	
3. Update the main branch to have an number of running instances at once
	
	This is where an **Override** can be useful, if we want the main branch to have an number of unlimited counters:
	```
	# POST /v1/limiter/rules/limit%20branch%20builds/overrides
	{
		"Spec": {
			"DBDefinition": {
				"Name": "override main branch",
				"GroupByKeyValues": {
					"repo_name": { "Type": 13, "Data": "willow"},
					"branch_name": { "Type": 13, "Data": "main"}
				}
			},
			"Properties": {
				"Limit": -1, // NOTE: this means unlimited in the API
			}
		}
	}
	```
	
	Now, when the runners try and create a **Counter** for the Willow's branch, they are all passing:
	```
	# PUT /v1/limiter/counters
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"},
					"branch_name": {"Type": 13, "Data": "main"},
					"os": {"Type": 13, "Data": "windows"}
				}
			},
			"Properties": {
				"Counters": 1
			}
		}
	},
	
	# PUT /v1/limiter/counters
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"},
					"branch_name": {"Type": 13, "Data": "new_feature_72"},
					"os": {"Type": 13, "Data": "mac"} // change in the os
				}
			},
			"Properties": {
				"Counters": 1
			}
		}
	},
	
	# PUT /v1/limiter/counters
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"},
					"branch_name": {"Type": 13, "Data": "new_feature_72"},
					"os": {"Type": 13, "Data": "linux"} // change in the os
				}
			},
			"Properties": {
				"Counters": 1
			}
		}
	},
	```
	Now, one might ask why did we add a **Rule** of "Unlimited" (`"Limit": -1`) for the **Override**? Should we
	not set that to 4? That seems like a very bad operation since that is a hardware limitation and will never 
	exceed the actual number of physical instances. This would mean we can increase or decrease those as we want
	without conflicting with the defined **Rules**.

4. OS limitations

	So far we have defined Rules for the limitations for the git branches to run, but what if we also want to enforce OS 
	limitations? Personally, I would only like to run the OS "Windows" and "Mac" on the main branch as those are more limited
	and can do that with an additional **Rule** and **Overrides**.
	
	
	```Rule
	# POST /v1/limiter/rules
	{
		"Spec": {
			"DBDefinition": {
				"Name": "limit os builds",
				"GroupByKeyValues": {
					"os": {"Type": 1024 }, // 1024 means 'Any'
				}
			},
			"Properties": {
				"Limit": 0 // now nothing will run if they have the key 'os'
			}
		}
	}
	```
	
	```Overrides
	# POST /v1/limiter/rules/limit%20os%20builds
	{
		"Spec": {
			"DBDefinition": {
				"Name": "willow any branch os linux",
				"GroupByKeyValues": {
					"os": {"Type": 13, "Data": "linux"},
				}
			},
			"Properties": {
				"Limit": -1, // NOTE: this means unlimited in the API
			}
		}
	},
	
	# POST /v1/limiter/rules/limit%20os%20builds
	{
		"Spec": {
			"DBDefinition": {
				"Name": "willow main override os windows",
				"GroupByKeyValues": {
					"os": {"Type": 13, "Data": "windows"},
					"branch_name": {"Type": 13, "Data": "main"}
				}
			},
			"Properties": {
				"Limit": -1, // NOTE: this means unlimited in the API
			}
		}
	},
	
	# POST /v1/limiter/rules/limit%20os%20builds
	{
		"Spec": {
			"DBDefinition": {
				"Name": "willow main override os mac",
				"GroupByKeyValues": {
					"os": {"Type": 13, "Data": "mac"},
					"branch_name": {"Type": 13, "Data": "main"}
				}
			},
			"Properties": {
				"Limit": -1, // NOTE: this means unlimited in the API
			}
		}
	},
	
	```

Hopefully that provides an example on how arbitrary **Key + Value Pairs** can be enforced depending on how a user wants to group them together. 
