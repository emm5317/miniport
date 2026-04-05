# Security Policy

## Important Warning

MiniPort provides unauthenticated access to Docker container management by default. The Docker socket (`/var/run/docker.sock`) grants root-equivalent privileges on the host. **You must take steps to secure access.**

### Recommended Setup

1. **Enable built-in auth** — Set `MINIPORT_AUTH` to point to an auth file containing `username:sha256hex` entries:
   ```bash
   echo -n 'your-password' | sha256sum | cut -d' ' -f1
   # Add to auth file as: admin:<hash>
   ```

2. **Use a reverse proxy with TLS** — Place MiniPort behind Caddy, nginx, or similar with HTTPS termination.

3. **Bind to localhost** — The default bind address is `127.0.0.1`. Do not change this to `0.0.0.0` without authentication and a reverse proxy.

4. **Restrict Docker socket access** — Run MiniPort as a non-root user in the `docker` group.

## Reporting Vulnerabilities

If you discover a security vulnerability, please open a GitHub issue or contact the maintainer directly. Please do not disclose vulnerabilities publicly until a fix is available.

## Scope

MiniPort is designed for single-user homelab deployments. It is not intended for multi-tenant or public-facing use.
