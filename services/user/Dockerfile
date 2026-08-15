FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app ./services/user/cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=build /app /app
ENV ADDR=:8080 DATABASE_PATH=/data/app.db JWT_SECRET_PATH=/data/.jwt_secret
EXPOSE 8080
CMD ["/app"]
