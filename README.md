# Neko Blogging Engine (Go Version)

## Docker

```sh
docker build -t docker-blog-go .

docker run -it --rm --name blog-go -p 8000:80 docker-blog-go:latest
```

## TODO

- [ ] Parser for custom markup language
- [ ] Generator for static pages
- [ ] Save meta data to Postgres database
- [ ] RSS/Atom feeds
- [ ] Dockerfile
- [ ] Auto-deploy?
- [ ] Webmention
