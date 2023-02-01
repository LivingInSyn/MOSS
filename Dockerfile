FROM golang:1.19 as builder

WORKDIR /usr/src/moss

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./cmd .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o moss ./...
# RUN go build -v -o /usr/local/bin/moss ./...

# move the build from the builder to a debian minimal image
from alpine:latest
WORKDIR /root/
# setup gitleaks and git
RUN apk add git
ADD https://github.com/zricethezav/gitleaks/releases/download/v8.15.0/gitleaks_8.15.0_linux_x64.tar.gz ./
RUN tar -xzf gitleaks_8.15.0_linux_x64.tar.gz
RUN mv gitleaks /usr/local/bin/gitleaks
# copy files from the builder
COPY --from=builder /usr/src/moss/moss/moss /root/moss
# move the configs from the project in
RUN mkdir /root/configs
COPY ./configs ./configs/
# setup and execute
RUN chmod +x /root/moss
CMD ["/root/moss"]
