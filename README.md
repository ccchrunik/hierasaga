# Eventual Completion Mechanism for Microservices

This is the final project of the course 2022 NTUEE Fault-Tolerance Computing. 

**Note:** We write a simulator in Go to simulated distributed services because of the short duration of a semester. We envision a realistic architecture using containers and Kubernetes for deployment.

## Introduction

Our goal is to apply fault tolerance to must-complete distributed transactions in the microservice architecture. As software systems become increasingly complex, many of them turn toward distributed architecture. With the help of virtualization and containerization, microservice architecture is adopted by many systems. In microservices, data are partitioned into several business domains to form a loosely-coupled architecture that can be updated and deployed independently. This also increases the modularity of the entire system, leading to less maintainence effort and faster deployment speed. Although most operations can be done within the same business domain, it is hard to perform traditional ACID transactions because of the possibility of network partition as data become distributed across several databases. The distributed nature results in uncertain behavior of a transaction.

This problem is especially serious in some critical system, including e-commerce websites. For example, if a customer had already complete a payment process while the system encounter a failure before this payment is persisted and the record is completely lost, it could lead to severe damage to customer relationship. As more and more third-party service integration is used in today's websites, this problem could happen more frequently because most of them are not idempotent. To prevent this type of problems, it is crucial to design a fault-tolerance mechanism to deal with this situation.

Among the previous methods, all of them focused on data consistency and isolation, in order to avoid the unstable states of distributed systems while executing transactions to be read; however, these methods cannot deal with technical errors, they can only roll back transactions due to business errors, product out of stock, for example. We would like to design a fault tolerant system that can recover the system even in the face of techinal errors including power failures, network partitions, and so on. 

The developer could choose to avoid using global gransactions (GT) but using local transactions (LT) instead to improve concurrency and performance at the cost of data consistency. We simply provide a way to enforce the correctness of distributed transactions. Our purpose mainly focus on fault-tolerance instead of data consistency and isolation. However, we provide a draft of a design to maintain the causality of operations, which is discussed in the last section.

## Related Work

- [SagaMAS: A Software Framework for Distributed Transactions in the Microservice Architecture](https://ieeexplore.ieee.org/abstract/document/8645853)
    - Provides a good formalized transaction model
    - Use agent at each service to handle distributed transactions (semi-orchestrated)
- [Enhancing Saga Pattern for Distributed Transactions within a Microservices Architecture](https://www.mdpi.com/2076-3417/12/12/6242)
    - Data Isolation using quota queue and commit-sync service
- [Limits of Saga Pattern](https://www.ufried.com/blog/limits_of_saga_pattern/#fnref:2)
    - Only handle business error, not techinal error
- [Application-Level Locking](https://stackoverflow.com/questions/5436021/implementing-application-level-locking)
- [2PC*: a distributed transaction concurrency control protocol of multi-microservice based on cloud computing platform](https://link.springer.com/content/pdf/10.1186/s13677-020-00183-w.pdf) 
- [Event-based data-centric semantics for consistent data management in microservices](https://dl.acm.org/doi/abs/10.1145/3524860.3539807?casa_token=101CsEviepkAAAAA:ZJb00U-bd7XInIMk9O6-pyS8UHveeJAAaEroeWHedRF2l0UJAXyVcJvkKgmLNiPoFmFBV5Nqi60dXQM)

## Transaction Model

We define the transaction model as the following to simplify the reasoning process.

### Architecture

![architecture](./img/diagram-architecture.drawio.png) 

### Assumptions

- **Microservice architecture**
  - A gateway to receive requests and return them to users.
  - A set of services and corresponding databases, each services may have multiple instances.
  - A coordinator, also a message broker, with a queue and control instances to dispatch events, and implement consistency requirements including distributed concurrency control.
  - A central data store to store globally visible information that spans across multiple services (there can be several databases, but only shown one in the architecture picture), such as causality relationships.
  - **Note**: this architecture is also a kind of *decentralized data management* as actual data do not cross the boundary of each service. The central data store only stores control information used for transactions.
- **Event-based architecture**
  - Asynchronous execution to reduce failure surface. If we use synchronous style, all related service must be online simultaneously.
  - Message queue, message broker as storage and executor.
  - Messages are only passed through synchronous method like HTTP or gRPC in the basic unit (more details later).
  - The communication between each component is through an **event**: a set of key-value pairs, usually in the form of json, xml, or protocol buffer.
  - **Note:** for APIs directly contacting each single service, they could either be implemented synchronously or asynchronously.
- **Heterogeneous environment**
  - RDB, KV store, Document DB, Time series DB, Event Store...
  - Bare-bone machine, virtual machine, containers, ...
- **Database Constraints**
  - All databases are durable, i.e. once a data is **commited** to the database, it will not disappeared on failures.
  - All local transactions performed by databases are **atomic** (*all-or-nothing* atomicity).
  - Some database may not support transactions, but all single operation in any database are **atomic** (*all-or-nothing* and *before-or-after* atomicity). One can relax this constraints but it will cause some degree of inconsistency.
  - **Note:** the developer **must** choose a database that support transactions or it could lead to some level of inconsistency.
- **Default configuration**
  - The location of the message broker, central data store (only visible to the coordinator), ...
  - Default retry configuration, the number of instances, ... 

### Requirements

- Each request that utilizes this mechanism (global transaction, GT) is gauranteed to complete all of the operation. However, for read/write that happens only outside of this mechanism (local transaction, LT) may cause certain degree of inconsistency. This is reasonable in order to improve performance and prevent blocking of the database in network partition.
- The execution time of the each request is bounded, usually within several hours.
- Allow rollback to a certain point when facing errors. If this transaction is commited, it would not allow rollback. That is, a transaction call rollback before the commit point, or it must proceed until the end.
  - The integration with third-party services or non-idempotent actions usually cannot rollback or it will lead to unexpected consequences.
- Allow intervention of developer (to manipulate the state of the execution context).
- Need the ability to deal with several types of errors
  - Business errors: the program logic is correct, but an user performed an illegal operation.
  - Techinical errors: the program logic is still corrects, but it will subject to the interaction between components and environments, e.g. network partition, power failure, memory killer, ...
  - Logical errors: the program is buggy, e.g. infinite loop.
- Prefer transparent to each service (by using agent, middleware, or application framework).
- Prefer easier programming style.

### Execution Model

- A tree of nodes starts from the root node.

![tree](./img/diagram-tree.drawio.png) 

- A global transaction can be composed of nonterminal or terminal transaction (local transaction).
- A nonterminal node can be described as a sequence of terminal nodes

![flatten](./img/diagram-flatten.drawio.png) 

- A terminal node can be described as a tuple of `(Service, Endpoint, Action, Stage, Input, Output, Next Service)`.
    - `(Payment, /v1/payment, retry, 3, input {}, output {}, Order)`


### Execution Context

- Transaction ID
- Start Time
- End Time (optional)
- Current retry Time
- Time to Live (total retry time + service time)
- Trace stack (stages had been executed)
- Rollback stack (compensation stages)
    - [payment/v1/rollback, order/v2/order/rollback]
- Next Stage (Service, Endpoint, Action, Version, Stage, Input)
- State (`BEGIN`, `PROCESSING`, `COMMIT`, `ABORT`, `END`)
- Action (`NONE`, `CHECKPOINT`, `ROLLBACK`, `RETRY`, `INSPECT`)
- [custome fields ...] (optional)

```json
{
  "transaction_id": TX123456789,
  "start_time": XXX,
  "end_time": YYY,
  "retry_time": 3,
  "ttl": 50,
  "trace_stack": ["Payment:/v1/payment", "Order:/v1/order"],
  "rollback_stack": ["Payment:/v1/payment/rollback", "Order:/v1/order/rollback"],
  "next_stage": "Customer:/v1/customer:1",
  "state": "COMMIT",
  "action": "CHECKPOINT",
  "input": {
  	...: ...
  },
  "custom_config": {
  	...: ...
  }
}
```

### Execution Behavior

- The root generates a transaction id and add start time.
- The final node make the end time.
- A nonterminal node can mark the transaction as commit or abort, then the coordinator would do the corresponding processing.
- Each node take the input from the context and output the next stage information including the output data
- Each node add the stage information (Service, Endpoint, Action, Stage) to the trace stack.
- Each node add the rollback information (Service, Endpoint, Action, Stage) to the rollback stack.
- Each node decrease the TTL by 1 after processing.
- If the exection failed as this stage, the retry time is decreased by 1. If success, then mark current retry time as -1.
- If the node see the current retry time is -1, then mark the current retry time as the default retry time for that service.
- If either TTL or retry time becomes 0, or any error occurs (rollback or cancel), the execution ends.

## Mechanism

In this section, we would discuss a mechanism step-by-step to deal with the possible error in the fault model in detail. 

**Note:** 

- In our model, the distributed transaction is implemented through event-based systems. The user will receive acknowledgement after our system has confirmed that the request is save in our data store. The request is executed asynchronously with best-effort.
- The **central data store** is not a off-the-shelf database but a specialized database which has transaction support, read/write coherence management, and some other features. We provide a draft for such a design. We could also use a database with ACID support to provided a subset of the fault-tolerance in our context.

### Fault Model

This fault model mainly deals with the faults from the environments of the interactions between components. The logic error of the execution flaws are not considered. This type of error need to be addressed by the developer itself or through some verification processes like testing.

There are 3 types of errors:

- Business Error:
  - This type of error is valid from the viewpoint of our systems. It can be viewed as a non-error transaction with the business rollback stage as a non-terminal node. We would not discuss this type of error in the context.
- Technical Error:
  - Service Failure: the service is temporarily shutdown or crashed.
  - Network Partition: the network connection failed when communicating with another components.
  - We are able to distinguish these two conditions because network partition usually result in timeout error rather than not-found error. We mainly deal with this type of error in our system.
- Logical Error:
  - The most notable error in this category is programming bugs like infinite loop.
  - This type of error is only mentioned a little bit during out discussion.

### Basic Unit

We can decompose the operation sequence into basic units, made up of sender and receiver. We deal with communication error in this section and assume that each side has the ability to deal with errors. There are 4 possibilities that the basic unit can fail.

![basic](./img/diagram-basic.drawio.png)

**Normal Operation:** The sender first does some works and then send a message to the receiver. The receiver processed the event and sends a response back to the sender.

(1) The sender fails to process the event before the message is sent to the receiver. The sender have to design its own mechanism to handle the error while retrying the request.

(2) While the communication breaks (either each side failed or network partition), it behaves like (1). Just retry.

(3) The receiver receives the message but the processing failed in the middle. In this case, the receiver must save the state of processing (if not idempotent) in terms of the next retry request.

(4) The processing completes but the acknowledgement fails in the response (either each side failed or network partition). It behaves like (3). However, the request is already completed. The receiver must design a mechanism to deal with duplicate request.

By formulating these type of request, we can simplify our reasoning in the following section.

**Note:** The only difference between the two paragraph is that the event is pushed or pulled from the source. The fault-tolerance mechanism is actually equivalent.

### Prepare Phase

There must always be a prepare phase before the start of a transaction. This phase conduct some works to enable failure recovery. In theory, this phase can combine with the *Begin Phase* but it will complicate the handling so we decide to separate them to two different stages.

![prepare](./img/diagram-prepare.drawio.png)

**Normal Operation:** When a service endpoint receive a request, it first issues a local transaction to store the needed information in the state store. Then it directly sends acknowledgement back to the user. The background thread of the service will periodically pull the state store and send a `BEGIN` event with necessary transaction information (see *Execution Context* in the section *Transaction Model*) to the message queue that can be retrieved by the coordinator.

(1), (2) The service issues a transaction to save the necessary information (indexed by transaction ID). Since it leverage the local transaction, it is safe to retry the transaction on failures. In these steps, we just retry the same operation until success. Usually, this kind of operation is implemented either using read-then-insert or directly upsert command.

(3) The background thread fetchs all unsent events from the state store, marks them with `BEGIN` state and put them in the message queues. In production, there will be several instances that doing the same things. To prevent race condition, which leads to duplicate messages, we put the deduplication mechanism at the begin phase.

(4), (5) As each put in the message queue is atomic, just like transactions, we can treat (4), (5) as a whole. Just retry putting the event.

(6) After the event is sent to the message queue, the service marks the state of the transaction as complete. We can implement by directly delete this entry or just marked it as complete. If we implement removing strategy and we receive an error when removing a non-exist entry, we know that this is an duplicate event and can just ignore it. If (6) failed, then it will lead to duplicate events. We leave this problem to *Begin Phase*.

**Example:** After the user pays the order, the third-party service will send a request through the given callback, and it will stop sending until we acknowledge that we have already received it. The service first save all the needed information in the local state store. If something fails during this process, since the service has yet acknowledge to the third-party service, it is safe to retry this operation. 


### Begin Phase

In this phase, our mission is to deduplicate the same messages and coordinate the conflicting events that require have causality relationships.

![begin](./img/diagram-begin.drawio.png)

**Normal Operation:** The coordinator read an event from the message queue. If the event is a `BEGIN` event, then it create a new entry in the data store (issue a local transaction), examine the causality relationships and get a ticket of execution. The coordinator acknowledge the message.  When in this ticket turn, the data store send an event to the message queue. In terms of duplicate events, if the coordinator finds an entry in the data store (using transaction ID), it removes the event from the message queue. As each opeartion in the data store is atomic, this operation is correct.

(1) The coordinator pulls an event from the message queue. On failures, just retry.

(2), (3) The coordinator issues a transaction to the central data store. The central data store creates a new entry, resolves the causality relationship, and return the response to the coordinator. If the central store fails, since it implements atomic operation, it is safe to retry. If the coordinator fails or communication link breaks, because our use of transaction, it is also safe.

(4) The coordinator acknowledge the event and remove it. If failed, the coordinator will know that the entry has been created (transaction ID), it can just retry.

(5), (6) In this transaction turn, the central data store sends an event (`PROCESSING` state) to the message queue. If failed, because the message queue support atomic operation, just retry.


### Processing Phase

After the *Begin Phase*, the transaction enters a chain of *Processing Phase*. 

**Note:** The *Local Database* here actually means a non-idempotent external service, including a database or a third-party API. We just illustrate it as a database for convenience.

![processing](./img/diagram-processing.drawio.png)

**Normal Operation:** The coordinator pull an event from the message queue. If the event is in `PROCESSING` state, then the coordinator sends a request based on the given endpoint. After the virtual service completes the operation, the coordinator first composes a new event and sends it to the message queue and then acknowledge the old one.

(1) The coordinator failed to pull an event from the message queue (or failed before (2)). It can just retry. However, we have to deal with the duplicate retry if the change is commited to the database. As we have already coordinate the global transaction to eliminate read/write conflicts, we can include a read check for `TXID` field in the database entry to find whether this is a duplicate transaction or not. See *Discussion* for more description of database schema design.

(2) The communication link failed between coordinator and virtual service before making any permanent change. Like (1), just retry.

(3), (4) Because of the local transaction, the virtual service is safe to retry.

(5) The virutal failed to send the response back to the coordinator due to some failures. As the change is commited, it could either retry the entire operation or rollback. We decide to retry and entire operation and perform deduplication. Please see (1).

(6) This situation is similar to (5) because both components are stateless. If the new event is sent but the old one is failed, it doesn't matter as we are safe to retry the request. See (1). Another possibility is that no event had been sent because of failures of the coordinator. In this case, the virtual service has to make event with action `INSPECT`. The coordinator will then look directly at the message queue by querying all the messages by the transaction ID to verify if the stage match the current state. If the stage is matched, meaning no new event had been sent, simply composes a new event and acknowledge the old one. If not match, meaning this event is a duplicate, just ignore it and acknowledge all old events. (If implemented correctly, there will only be 1 old event.)

**Note:** the *virtual service* here means a logical unit that has only one local transaction. If an endpoint originally initiates many non-idempotent change, including a local transaction, a non-reversible third-party API, and the like, it should be divided into multiple virtual services. This could be implemented using application framework. See *Simulation* for more description.

### Commit Phase

The *Commit Phase*  is similar to the *Processing Phase*. The only difference is that when seeing a `COMMIT` state and `CHECKPOINT` action in the metadata (see *Execution Context*), it will release the locks of the central data store (see *Discussion*).

![commit](./img/diagram-commit.drawio.png)

**Normal Operation:** The basic operation is the same as the *Processing Phase*. On return, if the state is `COMMIT` and the action is `CHECKPOINT`, the coordinator will clean up the resources before the commit point. This is equivalent to release the lock. The coordinator also composes a new event with the clean state and send the message to the message queue.


(1), (2), (3), (4), (5), (8) See the *Processing Phase*.

(6), (7) With local transaction, we are safe to retry this operation. Also, the central data store also stores the latest commit point so if the duplicate transaction arrives, it can return an error message so that the coordinator can safely discard this event.

**Note:** 

- When to commit? The commits points are marked by the **control endpoint**, which is actually the first endpoint. The initial endpoint is not only an entry point of the transaction, but also serves as the control of the service request and response. See *Discussion* for more information.
- There are many other actions that can happen in the commit stage:
  - When the action is `NONE`, it only signifies the transaction is committed.
  - When the action is `ROLLBACK`, which means the normal operations suffers from severe failures. The coordinator then starts rollback until the previous commit point.
  - When the action is `RETRY`, just retry this operation. (Can simply discard the response and let the entire operation pulled from the message queue.)

### Abort Phase

The *Abort Phase* is similar to both *Processing Phase* and *Commit Phase*. If a user issues a request that forces the transaction to stop, then it could abort the transaction and reverses all the change. After rollback operations, the transaction will go to the *End Phase*.

**Note:** The abort operation can happen due to programming error programming errors (like infinite loop) or explicit abort request from the user. The former one is contained in the metadata of each event and the latter one is stored in the central data store. If the coordinator see such abort request in the metadata, it can abort the operation. On the other hand, to improve performance, there is a background thread pulling data from the central data store and store them in the memory to track if some have abort this transaction. The coordinator will also abort this transaction when seeing the flag. See *Discussion* for more explanation.

![abort](./img/diagram-abort.drawio.png)

**Normal Operation:** The basic operation is the same as the *Processing Phase*. On return, if the state is `PROCESSING`, the coordinator looks at the cache in the memory to see if this transaction has been aborted. The other case is when the state of the event is set to `ABORT` (from virtual service). Both of the cases make the state of the transaction into `ABORT` and the coordinator changes the metadata of the event. It first pops the top of the `rollback_stack` and set the next stage to the rollback endpoint. On each rollback stage completion, it pops another entry in the stack until the stack is empty. This is just like what the *Processing Phase* does but with the stage endpoint subsituted by the rollback endpoint. When the stack is empty, the coordinator set the state to `END` and enters the *End Phase*.

(1), (2), (3), (4), (5), (7) See the *Processing Phase*.

(6) If the background thread had not updated the status of the event, it fine. It simply waits for the next round. If the checking failed, because this component is stateless, just retry.

**Note:** If the event has been commited, even if the central data store show the as `ABORT`. To explicitly kill the transaction (which is issued by the administrator), the central data store must show the state as `KILLED` rather than `ABORT`.

### End Phase

Just before the transaction end, there will be an `END` event sent by the last service. The coordinator will clean up the resource and complete the transaction when encounter this type of event.

![end](./img/diagram-end.drawio.png)

**Normal Operation:** The coordinator pulls an event from the message queue and find that this is an `END` event. It issues a cleanup request to the central data store to free allocated resources and then also cleanup the event in the message queue.

(1) The coordinator failed to pull the event from the message queue. This case has no effect, and the coordinator will keep pulling until success.

(2) The cleanup event is lost during the communication. It behave like (1), just retry as the coordinator only maintain no state in memory and is idempotent.

(3) The cleanup response is lost. The coordinator will retry the clean up request. Once success, it will receive an not-exist error message in response. The coordinator thus knows that the resources had already been cleaned up and proceeds to acknowledge the event.

(4) The communication breaks when transmitting the acknowledge message. It just behaves like (3) and retry until the event is acknowledged and removed from the queue.


## Simulation

### Implementation

```go
type Payment struct {
    
}
```

```go
type Order struct {
    
}
```

```go
type Customer struct {
    
}
```

### More Realistic Architecture

TODO

## Discussion

### Implicit Assumption

### Central Data Store

### 


