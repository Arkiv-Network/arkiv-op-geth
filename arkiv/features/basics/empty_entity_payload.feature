Feature: Empty Entity Payload

  Scenario: Empty entity payload
    When I submit an arkiv transaction with empty content and one annotation
    Then the transaction should succeed
    