Authorized Key + Values
-----------------------

## Current issue

`Willow` uses the Limiter to to setup a Rule + Override to ensure that the max number of Items queued does not go over the
configured Queue's limit. To do this, it uses `_willow_[name]` naming convention for Counters and Rule + Override names.
These need to be guarded against anyone else other than Admins (to perform emergency fixes for unforeseen errors) with editing them.

## Solution

Currently I propose that when clients are fully authorized, any Keys or other naming conventions that begin with an `_`
have to be fully 'service authorized' so that 3rd party users using the same services don't interfere with the requirements
of other services like `Willow`. Also having the `_[service]` naming convention makes things easy to see where the Counters/Rules
are coming from

Requirements:
  * need authorization (RBAC) setup before really diving deeper into this. Hopefully that will drive out how to setup the clients
    authorization for services vs users