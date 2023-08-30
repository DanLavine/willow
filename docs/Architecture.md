Architecture
------------

Willow can be broken down into 3 major logical components that split certain responsibilites


# Willow API server

Willow API Server is the main entry point for all clients and perform all operations for any Queues.
The server handles any logic pertaining to:
1. CRUD operations for Queues
2. CRUD operations for insert (Producer) or retrieve (Consumer) operations for Items in a queue
3. Authorization for clients connecting to the queue (TODO)

# Limiter API server

Limiter API Server acts as the main entry point for basic Limiter operations for the administrator.
These include:
1. CRUD operation for "rules" for the limiter

# OperationDB

The OperationDB is the heart of Willow and contains most of the logic for this project. The OperationDB is a wrapper
around the `btree_associated` data structure and provides a way of providing generic plugins to handle the logic
for any operations that make up a specific set of "tags".

Each plugin listed below can be thought of a "table" in ther DBs like MySQL, PSQL, MongoDB, etc. But instead
of a collection of rows that define a data set, Plugin DB treats a collection of "rows" as a grouping for a
logical operation to perform. So the rows are not the data itself, but an identifier to the logical operation.

SQL Example:
Consider a table in SQL DB like so:
```
CREATE TABLE Person (
    name string,
    age int,
)
```
This simple person contains all the information of a person and then its up to the application layer to
do something with that info when it is retrieved.

OperationDB Example:
The same Table can exist in OperatonDB, but now Person is a custom executable to run. In addition this,
the rows are are now any number of unique KeyValues.
``` Psudo Code
INSERT INTO Person (
  name "Dan",
  age 32,
)


INSERT INTO Person (
  name "brandy",
  age 25,
  children 2,
)

INSERT INTO Person (
  name "mac",
  age 23,
  location USA,
  star_sign "aries",
)
```
These are all valid entries to the Person's table as the rows are all unique. But now the Person's plugin can perform
an operation when a "person" is added to the table that can make use of, or totally ignore the rows as those just identify
what makes up a person, but not what to do with a person
