FROM golang:1.27.4-trixie AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /splitit ./cmd/splitit


FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /splitit /usr/local/bin/splitit

USER 10001:10001
EXPOSE 8080 9090

ENTRYPOINT ["splitit"]
CMD ["serve"]
