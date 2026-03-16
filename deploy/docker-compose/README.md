# Docker Compose Deployment Support

* [platform](platform) provides utilities which are expected to be used in development.
* [dc1](dc1) single data center deployment of the system

## Services

| Compose Project | Service  | Port  | Container Port | Description                   |
|-----------------|----------|-------|----------------|-------------------------------|
| dc1             | pg       | 16004 | 5432           | Dev instance Postgres         |
| platform        | lgtm     | 16000 | 3000           | Grafana                       |
| platform        | lgtm     | 16001 | 4317           | OTEL gRPC                     |
| platform        | lgtm     | 16002 | 4318           | OTEL HTTP/protobuf            |
| platform        | integ pg | 16003 | 5432           | Integration Postgres Instance |
