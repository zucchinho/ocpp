## Background

As a charge point operator (CPO), E-Flux provides an implementation of an Open Charge Point Protocol (OCPP) server which is configured on each of the charging stations that we manage.

This OCPP server enables two-way communication between a \*central management system (CMS) and the charge station itself for things like authorizing users to charge their vehicle at the station, configuration the charging station, and understanding the current state of the station. The communication is asynchronous.

- A charge station can have 1 or more connectors (read: plugs)
- Similar to a traditional energy meter, a charge station measures the amount of energy used based on the kilowatt hours between the start and end of a charging session per connector.

## Exercise

Given the events we have from the charge station, create a projection to be able to answer the following questions:

- How many charging stations are there?
- How many connectors does each charging station have?
- Is our current view of the number of charging station connectors correct?
- The current meter read value for each connector on the station?

## Input

Below is a link to an unordered list of events exchanged between a CMS and the OCPP server. Some events are ad-hoc (a station will periodically send them to us), and others follow the Request-Response pattern.

- Each event has:
- a `correlationId` - This is used to group related events. For example, the response to a request can be identified by checking values that share a `correlationId`.
- `messageType` identifies the type of the event.
  `occurredAt` - This is the time that the event took place. Events are not guaranteed to be in order
- `payload` - The payload unique to a specific event type.

### Event types

- `MeterValuesRequest` - Indicates that the CMS has requested up to date meter values from the charging station, for a given connector.
- `MeterValuesResponse` - Indicates that the charging station has responded to a `MeterValuesRequest`.
- `MeterValuesNotification` - An ad-hoc notification from the charging station notifying us of the current meter value of all of its connectors.
- `ConectorListRequest` Indicates that the CMS has requested information on the number of connectors that the charging station has.
- `ConnectorListResponse` - Indicates that the charging station has responded to a `ConectorListRequest`

https://gist.github.com/ggordan/152ff8db4edded58d32377d720e7b296
