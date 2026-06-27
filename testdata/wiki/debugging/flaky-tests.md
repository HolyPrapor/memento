# Flaky Tests Debugging

Known flaky tests and their root causes.

## Storage Migration Race

The indexer tests became flaky after the storage migration because the test
harness was not properly resetting the database between test cases. Each test
expected a clean state, but concurrent test runs would leave residual data
that caused assertion failures. The root cause is related to the
[Migration Steps](decisions/storage-v2.md#migration-steps) procedure.

Fix: Added a database reset in the test setup function that runs before each
test case.

## Network Timeout

HTTP client tests occasionally fail with timeout errors when run in CI. This is
caused by the CI environment's network configuration limiting outbound
connections. The fix was to use httptest servers for all HTTP-dependent tests.

## Random Seed Variability

Tests that use randomized data produce different results on different machines
because the seed is not fixed. This was resolved by using a fixed seed in test
setups and documenting the expected output ranges.
