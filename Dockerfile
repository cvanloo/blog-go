FROM golang:1.23.3-alpine3.20 AS build
#RUN apk add imagemagick ffmpeg libjxl libavif libheif
#RUN magick --version
#RUN magick -list format
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY cmd ./cmd
COPY assert ./assert
COPY markup ./markup
COPY page ./page
COPY config ./config
COPY public ./public
COPY 日記 ./日記
#RUN MAKE_ASSETS=1 go run ./cmd/koneko/koneko.go -env ./日記/.env -source ./日記 -out ./public
COPY ./日記/assets ./public/assets
RUN go run ./cmd/koneko/koneko.go -env ./日記/.env -source ./日記 -out ./public || :
RUN ls -lAhF public
RUN ls -lAhF public/assets

FROM caddy:2.9-alpine
COPY --from=build /usr/src/app/public/ /srv
COPY 日記/about.html /srv/about.html
COPY Caddyfile /etc/caddy/Caddyfile
