# alertme

**alertme** is a lightweight client-server metric collection service written in Go. It allows an agent to collect system metrics and send them to a central server for storage and retrieval.

## Features
- Collects system metrics at configurable intervals
- Stores metrics in a structured format
- Provides an HTTP API for updating and retrieving metrics
- Simple deployment with standalone binaries

## Getting Started

### Running the Server

1) Build the server binary:
    ```sh
    go build -o server cmd/server/main.go
    ```

2) Start the server:
    ```sh
    ./server -a localhost:8080
    ```
    - `-a` specifies the server address (default: `localhost:8080`).

### Running the Agent

1) Build the agent binary:
    ```sh
    go build -o agent cmd/agent/main.go
    ```

2) Start the agent:
    ```sh
    ./agent -a localhost:8080 -r 10 -p 2
    ```
    - `-a` specifies the server address.
    - `-r` sets the report interval in seconds.
    - `-p` sets the metric collection interval in seconds.

---

## API Reference

### List Metrics
**Endpoint:** `GET /`

**Description:** Returns an HTML page with a table of all stored metrics.

---

### Retrieve a Metric
**Endpoint:** `GET /value/{metricType}/{metricName}`

**Description:** Retrieves the value of a specific metric.

**Path Parameters:**
- `metricType` (string): The type of metric (`gauge` or `counter`).
- `metricName` (string): The name of the metric.

**Response:**
- `200 OK`: Metric value returned in the response body.
- `404 Not Found`: Metric does not exist.
- `400 Bad Request`: Invalid metric type.

**Example Request:**
```sh
curl -X GET http://localhost:8080/value/gauge/cpu_usage
```

**Example Response:**
```
45.3
```

---

### Update a Metric
**Endpoint:** `POST /update/{metricType}/{metricName}/{metricValue}`

**Description:** Updates or creates a metric with the specified value.

**Path Parameters:**
- `metricType` (string): The type of metric (`gauge` or `counter`).
- `metricName` (string): The name of the metric.
- `metricValue` (number): The value to assign to the metric.

**Response:**
- `200 OK`: Metric successfully updated.
- `400 Bad Request`: Invalid metric type or value.

**Example Request:**
```sh
curl -X POST http://localhost:8080/update/gauge/cpu_usage/45.3
```

---

## Tests

**Running external tests:** `iter1 -> iter5`

```shell
metricstest \
-test.run="^TestIteration([1-6]|[1-6][A-Z])$" \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-source-path=. \
-server-port=9000
```

**Running statictests:** 

```shell
go vet -vettool=`which statictest` ./...
```
