# Unified edge TLS (dev)

Generate a self-signed pair for local `secure-unified` bring-up:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt -subj '/CN=veil.local'
```

Production: mount real certs via `UNIFIED_NGINX_TLS_CERT` / `UNIFIED_NGINX_TLS_KEY` (see [docs/platform-unified-access.md](../../../docs/platform-unified-access.md)).
