# The image of the Go API (work area 2.3), after decktome:docker/api.Dockerfile.
# The Go version is the version of go/go.mod. Change the tag and the digest
# together. The digests were read from the registries on 2026-09-29.
FROM golang:1.27.1@sha256:3680233e3204827fbdc66088528ae6d4b3d034f51d03a99d454f6de034888244 AS build
WORKDIR /src
COPY go/ go/
# The deploy gives the commit, and GET /version returns it. The deploy
# check reads it to know which commit the live API runs.
ARG COMMIT=unknown
RUN cd go && CGO_ENABLED=0 go build -trimpath -ldflags "-X main.commit=${COMMIT}" -o /out/api ./cmd/api

# The nonroot image runs as uid 65532, with no shell. Cloud Run sets PORT.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
