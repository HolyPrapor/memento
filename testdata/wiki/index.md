# Test Wiki Index

This is the preamble for the index page. It contains general information about
the test wiki used for retrieval benchmarking.

## Getting Started

To get started with this project, follow the setup instructions. Make sure to
install all required dependencies and configure the environment properly. The
system uses a modular architecture with separate components for different tasks.

## Architecture Overview

The system is built around a core runner that orchestrates all operations. The
runner is separated from the scheduler to allow independent scaling of execution
and scheduling logic. This separation means the runner can handle heavy
workloads without blocking the scheduler from accepting new tasks.

## Design Decisions

Storage was migrated from flat files to SQLite in version 2. This change was
made to improve query performance and simplify the indexing pipeline. The
migration involved converting all existing data and updating the API layer.
