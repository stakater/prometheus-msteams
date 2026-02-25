# E2E Testing with Coverage

This directory contains end-to-end tests for prometheus-msteams.

## Running E2E Tests

### Basic E2E Tests (no coverage)
```bash
make test-e2e
```

### E2E Tests with Coverage
```bash
make test-e2e-coverage
```

### All Tests (Unit + E2E) with Merged Coverage
```bash
make test-all-coverage
```

This will:
1. Run all unit tests and collect coverage to `cover-unit.out`
2. Run all E2E tests and collect coverage to `cover-e2e.out`
3. Merge both coverage profiles into `cover.out`
4. Run the coverage threshold checks against the merged coverage
5. Clean up temporary coverage files

## Using Ginkgo/Gomega for E2E Tests

### Current Setup
The Ginkgo suite is currently commented out in `e2e_suite_test.go`. To enable it:

1. Uncomment the code in `e2e_suite_test.go`
2. Write your Ginkgo specs in the same directory
3. Run with Ginkgo CLI for better output:
   ```bash
   go run github.com/onsi/ginkgo/v2/ginkgo -v -coverpkg=github.com/stakater/prometheus-msteams/... -coverprofile=cover-e2e.out ./test/e2e/
   ```

### Ginkgo with Coverage Merge
To collect coverage from Ginkgo tests and merge with unit tests:

```bash
# Run unit tests with coverage
go test ./pkg/... -coverprofile cover-unit.out -covermode=atomic -coverpkg=./...

# Run Ginkgo E2E tests with coverage
go run github.com/onsi/ginkgo/v2/ginkgo -v \
  -coverpkg=github.com/stakater/prometheus-msteams/... \
  -coverprofile=cover-e2e.out \
  ./test/e2e/

# Merge coverage profiles
echo "mode: atomic" > cover.out
grep -h -v "^mode:" cover-unit.out cover-e2e.out >> cover.out

# Check coverage thresholds
./bin/go-test-coverage --config=./.github/.testcoverage-local.yml
```

## Example Ginkgo Test Structure

```go
var _ = Describe("Prometheus MSTeams E2E", func() {
    var (
        server *httptest.Server
        logger utility.Logger
    )

    BeforeEach(func() {
        logger = utility.NewLogger(utility.LogFormatJSON, false)
        // Setup test infrastructure
    })

    AfterEach(func() {
        if server != nil {
            server.Close()
        }
    })

    Context("When sending alerts to Teams webhook", func() {
        It("should successfully post O365 cards", func() {
            // Your test code here
            Expect(response.StatusCode).To(Equal(200))
        })

        It("should successfully post Workflow cards", func() {
            // Your test code here
            Expect(response.StatusCode).To(Equal(200))
        })
    })
})
```

## Integration Test Mode

Set the `INTEGRATION_TEST_WEBHOOK_URL` environment variable to run tests against a real Teams webhook:

```bash
INTEGRATION_TEST_WEBHOOK_URL="https://your-webhook-url" make test-e2e
```

## Coverage Tips

1. **Instrument code for E2E coverage**: The `-coverpkg` flag tells Go to collect coverage for specific packages even when tested from a different package.

2. **Use -covermode=atomic**: This is required when merging coverage profiles from parallel tests.

3. **View merged coverage**: After running `make test-all-coverage`, view the HTML report:
   ```bash
   go tool cover -html=cover.out
   ```

4. **Focus on critical paths**: E2E tests should focus on integration points and real-world scenarios that unit tests can't cover.

5. **Measure improvement**: Compare coverage before and after E2E tests:
   ```bash
   # Unit tests only
   go tool cover -func=cover-unit.out | tail -1
   
   # With E2E tests
   go tool cover -func=cover.out | tail -1
   ```
