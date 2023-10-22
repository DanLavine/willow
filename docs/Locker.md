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

1. On a server restart, what should happen to clients waiting for locks that are partialy thrrough a request?
    a. Should they be dropped and cleared from disk?
    b. Do we only wite to disk when we "have" all locks -> then respond -> shutdown
    c. Can we "clean up" partially written data, that was never stored?

2. Clients will need to re-establish which locks they actually have acces to using an "identifier"
    a. "identifier" -> datatypes.StringMap


## Usage