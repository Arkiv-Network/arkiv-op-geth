# Arkiv System Documentation

This document describes the transaction types and query RPC functionality of the Arkiv storage system.

## Transaction Types

Arkiv transactions support five operation types that can be combined atomically within a single transaction.

### 1. Create

Adds new entities to the storage layer.

**Fields:**
- `btl` (uint64): Blocks-to-live - number of blocks the entity will persist
- `contentType` (string): MIME type of the payload (max 128 chars)
- `payload` (bytes): The entity data content
- `stringAttributes` (array): Key-value pairs where values are strings
- `numericAttributes` (array): Key-value pairs where values are numbers

**Behavior:**
- Entity key is derived as: `keccak256(txHash, operationIndex)`
- Sets `owner` and `creator` to transaction sender
- Sets `expiresAtBlock` to `currentBlock + btl`
- Records `createdAtBlock`, `lastModifiedAtBlock`, `operationIndex`, `transactionIndex`

**Validation:**
- BTL must be > 0
- Content type required and ≤ 128 characters
- Attribute keys must match identifier regex
- No duplicate attribute keys within same type (string/numeric)

### 2. Update

Modifies existing entities by replacing content and attributes.

**Fields:**
- `entityKey` (hash): The key of the entity to update
- `btl` (uint64): New blocks-to-live value
- `contentType` (string): New MIME type (max 128 chars)
- `payload` (bytes): New entity data content
- `stringAttributes` (array): New string attributes (replaces all existing)
- `numericAttributes` (array): New numeric attributes (replaces all existing)

**Behavior:**
- Requires sender to be the entity owner
- Preserves original `owner`, `creator`, and `createdAtBlock`
- Updates `lastModifiedAtBlock`, `operationIndex`, `transactionIndex`
- Resets `expiresAtBlock` to `currentBlock + btl`

**Validation:**
- Same as Create (BTL > 0, content type required, attribute validation)
- Entity must exist
- Sender must be owner

### 3. Delete

Removes entities from storage.

**Fields:**
- Array of entity keys (hashes) to delete

**Behavior:**
- Requires sender to be the entity owner
- Permanently removes entity and all associated data
- Emits deletion event logs

**Validation:**
- Entity must exist
- Sender must be owner

### 4. Extend

Extends the time-to-live of an entity without modifying content or attributes.

**Fields:**
- `entityKey` (hash): The key of the entity to extend
- `numberOfBlocks` (uint64): Additional blocks to add to current expiration

**Behavior:**
- Does not require ownership (anyone can extend any entity)
- New expiration: `currentExpiresAtBlock + numberOfBlocks`
- Does not modify payload, attributes, or metadata

**Validation:**
- numberOfBlocks must be > 0
- Entity must exist

### 5. ChangeOwner

Transfers entity ownership to a new address.

**Fields:**
- `entityKey` (hash): The key of the entity
- `newOwner` (address): The new owner address

**Behavior:**
- Requires sender to be current owner
- Updates `owner` field to `newOwner`
- Preserves all other metadata including `creator`
- Emits ownership change event

**Validation:**
- Entity must exist
- Sender must be current owner

## Transaction Format & Compression

Transactions are encoded and compressed to minimize on-chain data:

1. **Encoding**: Transaction is RLP-encoded
2. **Compression**: Encoded data is Brotli-compressed
3. **Size Limit**: Maximum 20MB when decompressed
4. **Execution**: `UnpackArkivTransaction` decompresses and decodes calldata

The compressed transaction bytes are passed as calldata to the Arkiv processor contract.

## Transaction Semantics

### Atomicity

All operations within a transaction execute atomically - either all succeed or all fail. If any operation fails validation or execution, the entire transaction reverts.

### Event Logs

All Arkiv events are emitted from the Arkiv Processor contract address: `0x00000000000000000000000000000061726B6976`

Each operation type emits specific event logs as defined below.

#### ArkivEntityCreated

Emitted when a new entity is created via a Create operation.

**Event Signature**: `ArkivEntityCreated(uint256,address,uint256,uint256)`

**Topics**:
- `topics[0]`: Event signature hash (keccak256 of signature)
- `topics[1]`: Entity key (indexed)
- `topics[2]`: Owner address (indexed)

**Data** (64 bytes):
- Bytes 0-31: Expiration block number (uint256)
- Bytes 32-63: Cost in wei (uint256, currently always 0)

#### ArkivEntityUpdated

Emitted when an existing entity is modified via an Update operation.

**Event Signature**: `ArkivEntityUpdated(uint256,address,uint256,uint256,uint256)`

**Topics**:
- `topics[0]`: Event signature hash
- `topics[1]`: Entity key (indexed)
- `topics[2]`: Owner address (indexed)

**Data** (96 bytes):
- Bytes 0-31: Old expiration block number (uint256)
- Bytes 32-63: New expiration block number (uint256)
- Bytes 64-95: Cost in wei (uint256, currently always 0)

#### ArkivEntityDeleted

Emitted when an entity is explicitly deleted via a Delete operation.

**Event Signature**: `ArkivEntityDeleted(uint256,address)`

**Topics**:
- `topics[0]`: Event signature hash
- `topics[1]`: Entity key (indexed)
- `topics[2]`: Owner address (indexed)

**Data**: Empty (0 bytes)

#### ArkivEntityBTLExtended

Emitted when an entity's time-to-live is extended via an Extend operation.

**Event Signature**: `ArkivEntityBTLExtended(uint256,address,uint256,uint256,uint256)`

**Topics**:
- `topics[0]`: Event signature hash
- `topics[1]`: Entity key (indexed)
- `topics[2]`: Owner address (indexed)

**Data** (96 bytes):
- Bytes 0-31: Old expiration block number (uint256)
- Bytes 32-63: New expiration block number (uint256)
- Bytes 64-95: Cost in wei (uint256, currently always 0)

#### ArkivEntityOwnerChanged

Emitted when entity ownership is transferred via a ChangeOwner operation.

**Event Signature**: `ArkivEntityOwnerChanged(uint256,address,address)`

**Topics**:
- `topics[0]`: Event signature hash
- `topics[1]`: Entity key (indexed)
- `topics[2]`: Old owner address (indexed)
- `topics[3]`: New owner address (indexed)

**Data**: Empty (0 bytes)

#### ArkivEntityExpired

Emitted when an entity is automatically removed by the housekeeping system due to expiration.

**Event Signature**: `ArkivEntityExpired(uint256,address)`

**Topics**:
- `topics[0]`: Event signature hash
- `topics[1]`: Entity key (indexed)
- `topics[2]`: Owner address (indexed)

**Data**: Empty (0 bytes)

### Attributes

Attributes (formerly "annotations") are key-value metadata attached to entities:
- Used for indexing and querying
- Keys must match identifier regex: `^[a-zA-Z_][a-zA-Z0-9_]*$`
- Same key can have both string AND numeric values simultaneously
- Cannot have duplicate values of the same type

## Query RPC API

The Arkiv RPC API provides methods to query and retrieve entity data. Implementation is in [eth/api_arkiv.go](eth/api_arkiv.go).

### Query Method

`arkiv_query(queryExpression, options)` - Filter and retrieve entities matching a query expression.

**Parameters:**

1. `queryExpression` (string): Query expression parsed by the query engine
2. `options` (QueryOptions object):
   - `atBlock` (uint64, `0x` prefixed, hex encoded string): Historical block number for query
   - `includeData` (IncludeData object, optional): Which fields to return
   - `orderBy` (array, optional): Ordering by attributes
   - `resultsPerPage` (uint64, optional): Pagination limit
   - `cursor` (string, optional): Pagination cursor from previous response

**IncludeData fields:**
- `key` (bool): Entity key hash
- `attributes` (bool): User-defined attributes
- `syntheticAttributes` (bool): System-generated attributes ($sequence, $creator)
- `payload` (bool): Entity data content
- `contentType` (bool): MIME type
- `expiration` (bool): Expiration block number
- `owner` (bool): Owner address
- `createdAtBlock` (bool): Creation block number
- `lastModifiedAtBlock` (bool): Last modification block
- `transactionIndexInBlock` (bool): Transaction index
- `operationIndexInTransaction` (bool): Operation index


**Response:**
```json
{
  "blockNumber": 12345,
  "data": [ /* array of entity objects */ ],
  "cursor": "optional_pagination_cursor"
}
```

**Behavior:**
- Maximum response size: 512KB
- Returns cursor when response size limit or resultsPerPage reached
- Queries historical state when `atBlock` specified
- Waits briefly for future blocks (up to 2x block cadence)

### Helper Methods

#### GetEntityCount

`arkiv_getEntityCount()` - Returns total number of entities at current block.

**Returns:** `uint64` count of entities

#### GetNumberOfUsedSlots

`arkiv_getNumberOfUsedSlots()` - Returns storage slot usage accounting.

**Returns:** `hexutil.Big` number of used storage slots

#### GetBlockTiming

`arkiv_getBlockTiming()` - Returns current block timing information.

**Returns:**
```json
{
  "current_block": 12345,
  "current_block_time": 1699123456,
  "duration": 2
}
```

## Query Language

The Arkiv query system provides a powerful SQL-like language for filtering entities based on attributes and system metadata.

### Operators

**Equality & Inequality**:
- `=` - Equals
- `!=` - Not equals

**Comparison**:
- `<` - Less than
- `<=` - Less than or equal
- `>` - Greater than
- `>=` - Greater than or equal

**Pattern Matching**:
- `~` or `GLOB` - Glob pattern match (supports `*` and `?` wildcards)
- `!~` or `NOT GLOB` - Negated glob pattern match

**Set Membership**:
- `IN (value1 value2 ...)` - Value is in the set
- `NOT IN (value1 value2 ...)` - Value is not in the set

**Logical Operators**:
- `AND` or `&&` - Logical AND (case-insensitive)
- `OR` or `||` - Logical OR (case-insensitive)
- `NOT` or `!` - Logical negation

**Grouping**:
- `( )` - Parentheses for grouping and precedence control

### Special Attributes

System-defined attributes prefixed with `$`:

- `$key` - Entity key hash (hex string)
- `$owner` - Current owner address (hex string)
- `$creator` - Original creator address, immutable (hex string)
- `$expiration` - Expiration block number (uint64)
- `$sequence` - Modification sequence for ordering entities by modification time (uint64)
- `$all` - Special query matching all entities

### Value Types

- **Strings**: Must be quoted with double quotes: `"value"`
- **Numbers**: Unquoted integers: `123`
- **Addresses**: `0x` + 40 hex characters (can be quoted or unquoted): `0x1234567890123456789012345678901234567890`
- **Entity Keys**: `0x` + 64 hex characters (can be quoted or unquoted): `0xabcd...`

### Query Examples

**Match all entities**:
```
$all
```

**Basic equality**:
```
name = "test"
age = 123
status = "active"
```

**Inequality**:
```
status != "inactive"
type != "archived"
```

**Comparison operators**:
```
age > 18
price <= 100
```

**Logical AND** (both styles work):
```
name = "test" AND age = 30
name = "test" && age = 30
```

**Logical OR** (both styles work):
```
status = "active" OR status = "pending"
status = "active" || status = "pending"
```

**Negation**:
```
!(name = "test")
NOT (status = "inactive")
```

**GLOB pattern matching**:
```
name ~ "foo*"
name !~ "test*"
name GLOB "doc_*.pdf"
name NOT GLOB "temp*"
```

**IN operator**:
```
status IN ("active" "pending" "approved")
age IN (18 19 20 21)
priority IN (1 2 3)
```

**Special attribute queries**:
```
$owner = 0x1234567890123456789012345678901234567890
$key = 0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
$expiration > 1000
$creator = "0x1234567890123456789012345678901234567890"
$sequence > 500
```

**Complex queries with parentheses**:
```
(type = "document" OR type = "image") AND status = "approved"
(age >= 18 AND age <= 65) OR role = "admin"
!(deleted = true) AND (category = "public" OR $owner = 0x123...)
```

**Combining multiple conditions**:
```
name ~ "doc*" AND status = "active" AND $expiration > 5000
category IN ("news" "blog") AND published = true AND $owner = 0x123...
```

### Full RPC Query Example

Using `arkiv_query` with all options:

```javascript
// Query for active documents owned by a specific address
{
  "jsonrpc": "2.0",
  "method": "arkiv_query",
  "params": [
    "type = \"document\" AND status = \"active\" AND $owner = 0x1234567890123456789012345678901234567890",
    {
      "atBlock": 12345,
      "includeData": {
        "key": true,
        "attributes": true,
        "syntheticAttributes": true,
        "payload": true,
        "contentType": true,
        "expiration": true,
        "owner": true,
        "createdAtBlock": true,
        "lastModifiedAtBlock": false,
        "transactionIndexInBlock": false,
        "operationIndexInTransaction": false
      },
      "orderBy": [
        {
          "name": "$sequence",
          "type": "numeric",
          "desc": true
        },
        {
          "name": "priority",
          "type": "numeric",
          "desc": false
        }
      ],
      "resultsPerPage": 50,
      "cursor": null
    }
  ],
  "id": 1
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "blockNumber": 12345,
    "data": [
      {
        "key": "0xabcd...",
        "value": "0x68656c6c6f",
        "contentType": "application/json",
        "expiresAt": 15000,
        "owner": "0x1234567890123456789012345678901234567890",
        "createdAtBlock": 10000,
        "stringAttributes": [
          {"key": "type", "value": "document"},
          {"key": "status", "value": "active"}
        ],
        "numericAttributes": [
          {"key": "priority", "value": 1}
        ]
      }
    ],
    "cursor": "6a736f6e5f637572736f72..."
  },
  "id": 1
}
```

**Default columns** when `includeData` is `null` or omitted:
- `key`, `payload`, `contentType`, `expires_at`, `owner_address`, User-defined attributes

## Terminology Note

This system is transitioning to new domain language:
- **Use**: "attribute"
- **Deprecated**: "annotation"

Internal code may still reference "annotations" but all external documentation and APIs should use "attributes".
