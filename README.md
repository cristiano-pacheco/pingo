# Pingo

Pingo is a **self-hosted uptime monitoring REST API**.  
It provides reliable monitoring of HTTP endpoints, exposes results through a clear API, and can be integrated directly into your systems and workflows.  
When a monitored endpoint goes down, Pingo can send alerts via **email** or **webhook**.

---

## Features

- **HTTP Monitoring Management**
  - Create, read, update, and delete monitoring targets
  - Retrieve monitoring results for tracked endpoints
- **User Management**
  - User registration and account confirmation
  - Secure login with password and one-time password (OTP) verification
  - Authentication via **JWT tokens**
- **Alerting**
  - Configurable alerts via **email**  
  - Configurable alerts via **webhooks**

---

## Tech Stack

- [Golang](https://golang.org/) with [Fiber](https://gofiber.io/) — lightweight, high-performance web framework  
- [Postgres](https://www.postgresql.org/) — relational database  
- [Redis](https://redis.io/) — caching and session management  
- [GORM](https://gorm.io/) — ORM for Golang  
- [Kafka](https://kafka.apache.org/) — event streaming and asynchronous processing  
- [FX](https://uber-go.github.io/fx/) — dependency injection framework  
- [Cobra](https://github.com/spf13/cobra) — CLI framework  
- [Viper](https://github.com/spf13/viper) — configuration management  
- [Testify](https://github.com/stretchr/testify) — testing toolkit  
- [OpenTelemetry](https://opentelemetry.io/) — tracing and observability  
- [JWT](https://jwt.io/) — token-based authentication  
- [ZeroLog](https://github.com/rs/zerolog) - json log

---

## Getting Started

### Prerequisites
- Go 1.25+
- Docker (for Postgres, Redis, Kafka)

### Installation
```bash
git clone https://github.com/cristiano-pacheco/pingo.git
cd pingo
go mod tidy
```

## Make Commands

The project includes a **Makefile** to simplify common tasks:

### Development

* `make install-libs` – Install development tools (linters, mockery, swag, nilaway, vuln checker)
* `make run` – Run the API server
* `make migrate` – Run database migrations

### Code Quality

* `make lint` – Run `golangci-lint`
* `make vuln-check` – Run Go vulnerability check
* `make nilaway` – Run `nilaway` for nil checking
* `make static` – Run lint, vuln-check, and nilaway

### Testing

* `make test` – Run unit tests
* `make test-integration` – Run integration tests
* `make cover` – Run tests with coverage report (HTML output in `reports/cover.html`)

### Utilities

* `make update-mocks` – Regenerate mocks using `mockery`
* `make update-swagger` – Update Swagger documentation

## License

MIT License. See [LICENSE](./LICENSE) for details.