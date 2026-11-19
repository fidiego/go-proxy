# Go Proxy

Super simple http proxy.

| Component      | Description                         |
| -------------- | ----------------------------------- |
| `go-proxy`     | http proxy server                   |
| `go-proxy-cli` | cli interface for `go-proxy` server |

Example usage:

```bash
# proxy a GET request
curl localhost:8080/proxy\?url=https://www.example.com

# trace a url's redirect chain (returns an html partial used by the index page)
curl localhost:8080/redirect\?url=https://www.example.com
```

Open `http://localhost:8080/` for the web UI.

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
