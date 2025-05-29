Feature: index entity

  Scenario: index a single entity by string annotation
    Given I have enough funds to pay for the transaction
    When I store an entity with a string annotation
    Then I should be able to retrieve the entity by the string annotation

  Scenario: index a single entity by numerical annotation
    Given I have enough funds to pay for the transaction
    When I store an entity with a numerical annotation
    Then I should be able to retrieve the entity by the numeric annotation
