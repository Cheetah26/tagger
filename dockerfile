FROM debian:latest

RUN apt update && apt install -y fuse libwebkit2gtk-4.0-dev

COPY internal/app/build/bin/tagger /tagger

ENTRYPOINT ["/tagger", "--mount", "$DB", "$MOUNT"]
