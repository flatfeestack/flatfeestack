FROM alpine:3.19
RUN apk add --update --no-cache md4c
COPY bylaws.md .
RUN md2html bylaws.md -o bylaws.html
