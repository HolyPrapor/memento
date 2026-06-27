# Runner Architecture

The runner is the core execution engine responsible for processing tasks.

## Design Rationale

The runner is separated from the scheduler because the two components have
fundamentally different scaling requirements. The scheduler is lightweight and
needs to be highly available to accept incoming tasks, while the runner is
CPU-intensive and benefits from horizontal scaling across multiple machines.

This separation also simplifies testing — the runner can be tested in isolation
without mocking the full scheduling pipeline. The runner stores results using
the [Storage v2](decisions/storage-v2.md#why-sqlite) backend.

## Worker Pool

The runner maintains a configurable worker pool that processes tasks
concurrently. Each worker pulls from a shared queue and executes tasks
independently. The pool size should be tuned based on available CPU cores.

## Task Lifecycle

Tasks flow through the runner in several stages:
1. Dequeued from the scheduler's task queue
2. Assigned to an available worker
3. Execution with timeout handling
4. Result reporting back to the scheduler
