Feature: Write-ahead log

  Scenario: creating an entity
    Given I have enough funds to pay for the transaction
    When submit a transaction to create an entity
    Then the entity should be created
    And the write-ahead log for the create should be created

  Scenario: updating the payload of the entity
    Given I have created an entity
    When I submit a transaction to update the entity, changing the paylod
    Then the payload of the entity should be changed
    And the write-ahead log for the update should be created

  Scenario: deleting an existing entity
    Given I have created an entity
    When I submit a transaction to delete the entity
    Then the entity should be deleted
    And the write-ahead log for the delete should be created

  Scenario: deleting expired entities
    Given I have enough funds to pay for the transaction
    And there is an entity that will expire in the next block
    When there is a new block
    Then the expired entity should be deleted
    And the write-ahead log for the delete should be created
