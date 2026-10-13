FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/odds ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/odds /odds

LABEL org.opencontainers.image.title="DraftKings NFL odds"

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/odds"]
