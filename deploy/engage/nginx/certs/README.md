# TLS certificates for engage nginx (secure / pentest overlay)

Place `tls.crt` and `tls.key` here for local HTTPS testing, or set `ENGAGE_NGINX_TLS_CERT` / `ENGAGE_NGINX_TLS_KEY` in compose.

```bash
openssl req -x509 -nodes -days 2 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt -subj '/CN=veil-engage.local'
```

Do not commit production private keys.
