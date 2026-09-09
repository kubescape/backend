# Service Discovery

Service discovery defines the client & server behavior for discovery of a Kubescape-compatible Backend services.

In order to test your service discovery backend endpoint run:

```bash
go test ./... -url domain.example -version v1
```

```bash
go test ./... -url domain.example -version v2
```

```bash
go test ./... -url domain.example -version v3
```

v4 (adds the optional `otel-events` endpoint)

```
go test ./... -url domain.example -version v4
```