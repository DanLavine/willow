Willow
------
[api](https://danlavine.github.io/willow/docs/openapi/willow/)

# TERMS
<dl>
<dt>Key</dt>
  <dd>Is always defined by a string</dd>
<dt>Value</dt>
  <dd>Is an object that defines what the type is (I.E. int, int32, float64, string, etc) and the data for that type</dd>
<dt>Key + Value Pairs</dt>
  <dd>a collection of Key + Values</dd>
<dt>Queue</dt>
  <dd>Contains a number of Channels that are logically grouped together for commain api calls. They allow for Consumers
  to query for any number of channels at once and enable easy bulk api operations on channels</dd>
<dt>Channel</dt>
  <dd>Contains a list of Items that that are queued for processing. The Channel itself is defined as the unique 
  Key + Values Pairs</dt>
<dt>Item</dt>
  <dd>Data that is enqueued to the Willow service from an end user to be processed</dd>
<dt>Producer</dt>
  <dd>Remote client that is connected to Willow and will insert Items into the Queue's channels </dt>
<dt>Consumer</dt>
  <dd>Remote client that is connected to Willow and will query Queue's channels for Items to retrieve and process</dt>
</dl>

# Service Features

1. Willow is a message queue where any **Items** enqueued can be 'updated' in place as long as no **Consumers** have
   picked them up for processing.

2. Before any **Items** are dequeued, Willow will create a Counter for the Limiter service with all the **Key + Value Pairs**
   defined on the **Channel** the **Item** belongs to. This way 3rd party Rules can be setup in the Limiter to enforce
   custom dequeue limits. Willow will then monitor to know when the **Channel** is under the failed Limiter Rules and
   allow for dequeue operations to try again

3. When an **Item** is resolved with either a success or failure, Willow will also update the custom counters again for
   to ensure that everything is up to date. In addition to this, the Item can be dropped as part of the fail operation 
   when it might want to re-queue. This is because items can be 'updated' in place and if it wants to re-queue at the front
   of the **Channel**, but there is another Item not yet being processed. The **Item** will just be discarded.

4. Consumers can query a **Queue's** **Channels** for possible **Key + Value Pairs** they might be interested in.

# Consumer Query Example

If you have followed the docs from the Limiter service, then this builds off of the custom Rules to enforce build policies
around each branches only having 1 available build at a time. As well as OS restriction for Main to only run Windows and Mac
builds, while all branches can build Linux.

1. First create a queue to manage all our git commits that can possibly build

	We can first create a new queue responsible for pushing possible git Commits to that any CICD runner can pull from
	```
	# POST /v1/queues
	{
		"Spec": {
			"DBDefinition": {
				"Name": "git commits"
			},
			"Properties": {
				"MaxItems": 50  // across all channels, how many items can be enqueued. Setup with a Limiter Rule + Override
			}
		}
	}
	```

2. Next we can enqueue items to channels

	When Enqueuing an **Item** each **Channel** is created on the fly if it does not currently exist
	```
	# POST /v1/queues/git%20commits/channels/item
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"},
					"branch_name": {"Type": 13, "Data": "new_feature_72"},
					"os": {"Type": 13, "Data": "linux"},
				}
			},
			"Properties": {
				"Item": [git commit to build], // bytes
				"Updateable": true,            // if this is true and no item is processing. another request that comes in will overwrite the entire Item
				"RetryAttempts": 1,            // how many times to attempt a retry
				"RetryPosition": "front",      // if it fails and there is another item in the queue. This Item will be dropped as it is 'updated'
				"TimeoutDuration: 1000000000,  // 1 second represented as nanoseconds. Refreshed via heartbeats

			}
		}
	}
	```
	So in this case, the **Channel** is defined as the values for `{KeyValues: ...}`.
	
	Item number two:
	```
	# POST /v1/brokers/queues/git%20commits/channels/items
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"},
					"branch_name": {"Type": 13, "Data": "main"},
					"os": {"Type": 13, "Data": "windows"},
				}
			},
			"Properties": {
				"Item": [git commit to build], // bytes
				"Updateable": true,            // if this is true and no item is processing. another request that comes in will overwrite the entire Item
				"RetryAttempts": 0,            // how many times to attempt a retry
				"RetryPosition": "front",      // if it fails and there is another item in the queue. This Item will be dropped as it is 'updated'
				"TimeoutDuration: 1000000000,  // 1 second represented as nanoseconds. Refreshed via heartbeats

			}
		}
	}
	```
	
	Item number three:
	```
	# POST /v1/queues/git%20commits/channels/items
	{
		"Spec": {
			"DBDefinition": {
				"KeyValues": {
					"repo_name": {"Type": 13, "Data": "willow"},
					"branch_name": {"Type": 13, "Data": "main"},
					"os": {"Type": 13, "Data": "linux"},
				}
			},
			"Properties": {
				"Item": [git commit to build], // bytes
				"Updateable": true,            // if this is true and no item is processing. another request that comes in will overwrite the entire Item
				"RetryAttempts": 0,            // how many times to attempt a retry
				"RetryPosition": "front",      // if it fails and there is another item in the queue. This Item will be dropped as it is 'updated'
				"TimeoutDuration: 1000000000,  // 1 second represented as nanoseconds. Refreshed via heartbeats

			}
		}
	}
	```

3. Consumer trying to pull a value for the Linux runner

	Now a runner on the Linux runner can run a general query to pull any of the Branches that are on the Linux runner:
	```
	# GET /v1/queues/git%20commits/channels/items
	{
		"Selections": {
			"KeyValues": {
				"os": {
					"Value":{"Type": 13, "Data": "linux"},                     // specificaly look for any values with 'linux'
					"Comparison": "=",                                         // only use values that match exactly
					"TypeRestrictions": {"MinDataType": 13, "MaxDataType": 13} // ensure that when selecting keys, we don't select 'Any' saved in the DB
				}
			}
		}
	}
	```
	So in this case **Item** one or three can be dequeued because they both have the 'os' set to 'Linux'
	
	If however we wanted to ensure that we pull the main branch builds, we can add the desire key value
	```
	{
		"Selections": {
			"KeyValues": {
				"os": {
					"Value":{"Type": 13, "Data": "linux"},                     // specificaly look for any values with 'linux'
					"Comparison": "=",                                         // only use values that match exactly
					"TypeRestrictions": {"MinDataType": 13, "MaxDataType": 13} // ensure that when selecting keys, we don't select 'Any' saved in the DB
				},
				"branch_name" {
					"Value":{"Type": 13, "Data": "main"},                     
					"Comparison": "=",                                        
					"TypeRestrictions": {"MinDataType": 13, "MaxDataType": 13}
				}
			}
		}
	}
	```
	Now in this case, the Consumer is also enforcing that the branch_name is for the Main brach.
	
	In either case, Willow will check the Limiter to ensure our previous Rules are enforced by sending the **Channel's** 
	**Key + Value Pairs** as a Counter. In any case where a **Channel** has reached the Limits, the **Items** will not be
	dequeued and Willow will take care of resuming the **Channel** when Items are acknowledged by the clients or time out
	to decrement the Counters.
	
	The clients provided in the `pkg` folder will take care of automatically heartbeat this item for you.

4. Lastly we want to report the Item has finished processing

	Eventually the **Consumer** will need to respond to the server to acknowledge if the **Item** with a success or failure message.
	On success, the **Item** will be removed from the queue and the **Channel** will automatically be destroyed if there are no more
	**Items** currently enqueued. On a failure, the **Item** will automatically be re-queued if it is configure to do so. There are
	also additional rules if the **Item** is updatable and wants to re-queued at the "front". In that case the **Item** would be
	discarded since it would just be updated anyway. See the API docs for all the logical workflows.
	
	```
	# POST /v1/queues/git%20commits/channel/items/ack
	{
		"ItemID": "[guid]", // reported on the State, when dequing an item
		"Success": true,
		"KeyValues": {
			"repo_name": {"Type": 13, "Data": "willow"},
			"branch_name": {"Type": 13, "Data": "main"},
			"os": {"Type": 13, "Data": "linux"},
		}	
	}
	```
