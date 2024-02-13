Locker
------
[api](https://danlavine.github.io/willow/docs/openapi/locker/)


# Terms
<dl>
<dt>Key</dt>
  <dd>Is always defined by a string</dd>
<dt>Value</dt>
  <dd>Is an object that defines what the type is (I.E. int, int32, float64, string, etc) and the data for that type</dd>
<dt>Key + Value Pair</dt>
  <dd>a collection of Key + Values</dd>
<dt>Lock</dt>
  <dd>unique Key + Value Pairs that define a single lock</dd>
</dl>

# Service Features

Locker provides a generic locker service that can be used to **Locks** locks from arbitrary **Key + Value Pairs**. 
Currently the Locker Service generates only exclusive **Locks** for any request and must wait for a previous request to be
finished before obtaining the **Lock**. Then if there are no more clients waiting for a **Lock** it is removed from the
service to keep resource usage down.

Each **Lock** that is obtained must be heartbeat by the client to ensure that the client still maintains the **Lock**. If this
times out for any reason the service will release the **Lock** and allow for another client to obtain it. This means that the
clients must also be responsible for ensuring that if a **Lock** is lost, then they have to take appropriate actions. The clients
provided with this codebase should allow for time outs locally if they are unable to reach the Locker Service.  