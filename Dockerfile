FROM golang:1.19

WORKDIR /usr/src/moss

ADD https://github.com/zricethezav/gitleaks/releases/download/v8.15.0/gitleaks_8.15.0_linux_x64.tar.gz ./
RUN tar -xzf gitleaks_8.15.0_linux_x64.tar.gz
RUN mv gitleaks /usr/local/bin/gitleaks

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./cmd .
COPY ./configs .

RUN go build -v -o /usr/local/bin/moss ./...

CMD ["moss"]