Locker
------

Locker is a generic locking service for an arbitrary collection of Key Value pairs needs to be locked as a single unit

## API
TODO

(put swagger here)

## Logic
TODO

## Gurantees
TODO

## Open Questions

Should there be a "Match" api as well as a "Query" api. "Match" doesn't have much use in terms
of the locker service as the KeyValues need to be exact...

So in the Limiter service, "Match" is more of a "generic find anything that contains these key values".
So in that regard, "Match" is a behavior of the Limiter service.

## Usage