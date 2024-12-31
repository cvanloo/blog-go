# Neko Blogging Engine (Go Version)

## Docker

```sh
docker build -t docker-blog-go .

docker run -it --rm --name blog-go -p 8000:80 docker-blog-go:latest
```

## TODO

- [x] Parser for custom markup language
- [x] Generator for static pages
- [x] RSS/Atom feeds
- [x] Dockerfile
- [ ] Lexer+Parser: alternative syntax for double enquote: ``.....''
- [ ] Lexer+Parser: allow escaping of ) in links
- [ ] Auto-deploy?
- [ ] Webmention
