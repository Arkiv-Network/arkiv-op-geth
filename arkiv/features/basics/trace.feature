Feature: Tracing

  Scenario: tracing a storage transaction
    Given I have created an entity
    When I trace the transaction that created the entity
    Then the trace should be empty
