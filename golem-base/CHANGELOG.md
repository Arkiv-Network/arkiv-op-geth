# Changelog

Here we will keep track of the changes to the stock [go-ethereum](https://github.com/ethereum/go-ethereum) codebase.

## 2025-01-20
    - Forked the go-ethereum codebase at the tag v1.14.12 (commit `293a300d64be3d9a1c2cc92c26fcff4089deadcd`)
    - Added nix flake with DevShell for development

## 2025-01-22
    - Added UpdateStorageTx type and tests.

## 2025-01-24
    - Fixed transaction signing and processing of UpdateStorageTx.

## 2025-01-27
    - Added code to store golem base entities in the Account trie.

## 2025-01-28
    - Changed the semantics of UpdateStorageTx processing.

## 2025-01-30
    - Implemented end-to-end creation of entities.

## 2025-01-31
    - Added storing list of blocks that should expire by the block number (preparation for the entity TTL expiration functionality).

## 2025-02-03
    - Added first simple cucumber tests.
    - Added storing of annotation index (name->list of entities) for the exact search.


## 2025-02-04
    - Added simple query RPC: `golembase_searchEntitiesByAnnotations`
        Payload is of the form:
        ```json
            {
                "eq":{
                    "foo": "bar"
                }
            }
        ```
        where `eq` object is key/value pair of object annotations.

## 2025-02-05
    - Added deleting of entities through using UpdateStorageTx
    - Added updating of entities through using UpdateStorageTx

## 2025-02-06
    - Added `golembase_queryEntities` RPC method.
        Query language is of the form:
        ```
        (foo = "bar" || foo = "baz") && foo = "bar"
        ```
        where `foo` is the name of the annotation, and `bar` and `baz` are the values to be matched.
        The query language supports `AND` (`&&`) and `OR` (`||`) operators and parentheses for grouping.
        It also supports numeric annotations of the form `foo = 42`.

## 2025-02-12
    - Added a housekeeping transaction that expires entities that have exceeded their TTL.
      This transaction is added as a first transaction to the block.
      Processing of the housekeeping transaction deletes all entities that have exceeded their TTL and emits the corresponding events.

## 2025-02-14
    - Switched to using Account storage for storing entities.

## 2025-02-17
    - Rebased on top of Geth v1.15.1.

## 2025-02-18
    - Added writing of the write-ahead log to be picked up by ETLs.

## 2025-02-24
    - Ported code to op-geth

## 2025-03-03
    - Implemented ETL from WAL to SQLite

## 2025-03-10
    - Implemented ETL from WAL to MongoDB

## 2025-03-14
    - Implemented global list of entities

## 2025-03-19
    - Removed custom transactions to reduce fricion with the rest of the OP ecosystem

## 2025-03-24
    - Added storing of entity owners when entities are created
