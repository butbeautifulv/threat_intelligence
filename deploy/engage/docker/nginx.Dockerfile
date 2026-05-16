FROM nginx:1.27-alpine
RUN rm -f /etc/nginx/conf.d/default.conf
COPY deploy/engage/nginx/engage.conf /etc/nginx/conf.d/engage.conf
EXPOSE 443
