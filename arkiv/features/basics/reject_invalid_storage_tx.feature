Feature: Reject invalid storage transactions

  Scenario: Reject invalid storage transaction
    When I submit a storage transaction with no playload
    Then the transaction submission should fail

  Scenario: Reject invalid storage transaction
    When I submit a storage transaction with unparseable data
    Then the transaction should be rejected

  Scenario: Reject transactions with too many operations
    When I submit a storage transaction with 2001 operations
    Then last error should mention "number of operations is greater than 1000"

