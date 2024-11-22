FROM caddy:2.9-alpine
COPY public/ /srv
COPY Caddyfile /etc/caddy/Caddyfile
