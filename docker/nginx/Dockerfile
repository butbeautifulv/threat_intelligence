# syntax=docker/dockerfile:1
FROM nginx:1.26-alpine
COPY docker/nginx/nginx.conf /etc/nginx/nginx.conf
RUN mkdir -p /tmp/client_temp /tmp/proxy_temp /tmp/fastcgi_temp /tmp/uwsgi_temp /tmp/scgi_temp \
  && chown -R nginx:nginx /tmp/client_temp /tmp/proxy_temp /tmp/fastcgi_temp /tmp/uwsgi_temp /tmp/scgi_temp /var/cache/nginx /var/log/nginx \
  && chmod -R 755 /var/log/nginx
USER nginx
EXPOSE 8080
