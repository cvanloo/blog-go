FROM golang:1.23.3-alpine3.20 AS build
WORKDIR /usr/src/app
#COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download && go mod verify
COPY cmd ./cmd
COPY assert ./assert
COPY markup ./markup
COPY public ./public
COPY .env ./.env
RUN go run ./cmd/koneko/koneko.go generate -out ./public

FROM caddy:2.9-alpine
COPY --from=build /usr/src/app/public/ /srv
COPY Caddyfile /etc/caddy/Caddyfile
