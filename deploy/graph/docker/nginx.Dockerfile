# syntax=docker/dockerfile:1
FROM nginx:1.27-alpine
RUN rm -f /etc/nginx/conf.d/default.conf
COPY deploy/graph/nginx/upstreams.conf /etc/nginx/conf.d/00-upstreams.conf
COPY deploy/graph/nginx/security.conf /etc/nginx/conf.d/01-security.conf
COPY deploy/graph/nginx/veil.conf /etc/nginx/conf.d/veil.conf
RUN mkdir -p /etc/nginx/certs
EXPOSE 443
