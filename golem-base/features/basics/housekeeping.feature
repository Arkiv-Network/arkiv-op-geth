Feature: housekeeping
  Housekeeping transaction is automatically added as the first transaction in the block.
  It deletes expired entities from the state.
  
  Scenario: housekeeping transaction is automatically added as the first transaction in the block
    Given I have enough funds to pay for the transaction
    When there is a new block
    Then the housekeeping transaction should be submitted
    And the housekeeping transaction should be successful

  Scenario: deleting expired entities
    Given I have enough funds to pay for the transaction
    And there is an entity that will expire in the next block
    When there is a new block
    Then the expired entity should be deleted
    And the number of entities should be 0
    And the list of all entities should be empty
