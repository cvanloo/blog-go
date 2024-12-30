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
- [ ] Lexer: lex single quote as amp special, &rsquo; for apostrophe, e.g. in it's -> it&rsquo;s, can't -> can&rsquo;t
- [ ] Lexer: lex single enquote: `.....'
- [ ] Lexer: alternative syntax for double enquote: ``.....''
- [ ] Auto-deploy?
- [ ] Webmention
