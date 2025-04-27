# Go API with OpenTelemetry Demo

This repository contains a simple Go ToDo API demonstrating the integration of OpenTelemetry for distributed tracing, metrics, and logging. The setup uses Docker Compose to run the application alongside Jaeger, Prometheus, Grafana, Loki, and the OpenTelemetry Collector.

## Features

*   **ToDo API**: Basic CRUD operations for managing ToDo items (`/add`, `/list`, `/get`, `/complete`, `/delete`, `/update`, `/search`).
*   **OpenTelemetry Integration**:
    *   **Distributed Tracing**: Traces are generated for HTTP requests and exported to Jaeger via the OpenTelemetry Collector.
    *   **Metrics**: Application metrics (request latency, error counts, task counts) are exposed via Prometheus endpoint (`/metrics`) and collected by Prometheus via the OpenTelemetry Collector.
    *   **Logging**: Application logs are written to a file, collected by Promtail, and sent to Loki. Logs are correlated with traces using Trace IDs.
*   **Observability Stack**: Includes pre-configured Jaeger, Prometheus, Grafana (with basic dashboards/datasources), and Loki for visualizing telemetry data.

## Prerequisites

*   [Docker](https://docs.docker.com/get-docker/)
*   [Docker Compose](https://docs.docker.com/compose/install/)
*   [jq](https://stedolan.github.io/jq/download/) (for running the test script)
*   Go (v1.21 or later, only needed if you want to build/run outside Docker)

## Running the Application

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/DerfOh/todo-otel.git
    cd todo-otel/
    ```

2.  **Build and start the services using Docker Compose:**
    ```bash
    docker-compose up --build -d
    ```
    This command builds the `todo-app` image and starts all the services defined in `docker-compose.yml` in detached mode.

3.  **Verify services are running:**
    ```bash
    docker-compose ps
    ```
    You should see `jaeger`, `prometheus`, `grafana`, `loki`, `promtail`, `otel-collector`, and `todo-app` running.

## Testing the API

A simple test script is provided to interact with the API endpoints:

```bash
./test-http.sh
```

This script will:
*   List initial ToDos (should be empty).
*   Add several new ToDo items.
*   List ToDos again to show the added items.

Each request made by the script should generate traces viewable in Jaeger and update metrics in Prometheus.

## Accessing the Tools

*   **ToDo API**: `http://localhost:8080` (e.g., `http://localhost:8080/list`)
*   **Jaeger UI**: `http://localhost:16686` (Find traces for the `todo-app` service)
*   **Prometheus UI**: `http://localhost:9090` (Check targets and query metrics like `todo_handler_latency_milliseconds_bucket`, `todo_tasks_added_total`, `todo_handler_errors_total`)
*   **Grafana UI**: `http://localhost:3000` (Default login: `admin`/`admin`. Datasources for Prometheus, Jaeger, and Loki should be pre-configured)
*   **Loki** (via Grafana): Use the "Explore" view in Grafana and select the "Loki" datasource to query logs (e.g., `{job="docker"}`).

## Code Structure

The Go application code is organized as follows:

*   `main.go`: Entry point, sets up the HTTP server, initializes components, handles graceful shutdown.
*   `models.go`: Defines data structures (`ToDo`, `CompletedToDo`).
*   `store.go`: In-memory storage logic for ToDo items.
*   `handlers.go`: HTTP request handlers for API endpoints.
*   `telemetry.go`: OpenTelemetry initialization (tracing, metrics) and helper functions.
*   `logger.go`: Logging setup using `zerolog`, including file logging and rotation logic.
*   `utils.go`: Utility functions (e.g., `contains`).
*   `handlers_test.go`: Unit tests for HTTP handlers.

## Configuration

*   **`docker-compose.yml`**: Defines all services, ports, volumes, and networks.
*   **`otel-collector-config.yaml`**: Configures the OpenTelemetry Collector (receivers, exporters, pipelines).
*   **`prometheus.yml`**: Configures Prometheus scrape targets.
*   **`grafana/provisioning/`**: Contains Grafana datasource and dashboard provisioning files.
*   **`loki-config.yaml`**: Configuration for the Loki logging backend.
*   **`promtail-config.yaml`**: Configuration for Promtail log collection agent.

## Stopping the Application

```bash
docker-compose down
```
This command stops and removes the containers defined in the `docker-compose.yml` file. Add `-v` if you want to remove the volumes as well.
