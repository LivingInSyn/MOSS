FROM golang:1.19 as builder

WORKDIR /usr/src/moss

ADD https://github.com/zricethezav/gitleaks/releases/download/v8.15.0/gitleaks_8.15.0_linux_x64.tar.gz ./
RUN tar -xzf gitleaks_8.15.0_linux_x64.tar.gz
RUN mv gitleaks /usr/local/bin/gitleaks

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./cmd .
COPY ./configs .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o moss ./...
# RUN go build -v -o /usr/local/bin/moss ./...

# move the build from the builder to a debian minimal image
from alpine:latest
WORKDIR /root/
COPY --from=builder /usr/src/moss/moss /root/moss
RUN chmod +x /root/moss
CMD ["/root/moss"]
