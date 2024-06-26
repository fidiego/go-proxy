# Go Proxy

Super simple http proxy.

| Component      | Description                         |
| -------------- | ----------------------------------- |
| `go-proxy`     | http proxy server                   |
| `go-proxy-cli` | cli interface for `go-proxy` server |

Example usage:

```bash
curl localhost:8080/proxy\?url=https://www.example.com
```

## Development

### Requirements

| Tool     |                |
| -------- | -------------- |
| `just`   | Command Runner |
| `docker` | Containers     |

### Running in dev

```bash
just up
```

### Other Commands

Run `just` or `just default` to get a list of available commands.
