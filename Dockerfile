FROM ubuntu:latest

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
COPY mysqlsetup.sql /tmp

RUN apt-get update && apt-get install -y wget git gcc vim software-properties-common mysql-server && add-apt-repository -y ppa:ethereum/ethereum && apt-get update && apt-get install -y ethereum && wget -P /tmp https://go.dev/dl/go1.18.3.linux-amd64.tar.gz && tar -C /usr/local -xzf /tmp/go1.18.3.linux-amd64.tar.gz && rm /tmp/go1.18.3.linux-amd64.tar.gz && mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && cd $GOPATH/src && go mod init ethmod && go get -u github.com/ethereum/go-ethereum gorm.io/gorm gorm.io/driver/mysql github.com/gin-gonic/gin && go mod tidy && service mysql start && mysql < /tmp/mysqlsetup.sql && service mysql stop && rm /tmp/mysqlsetup.sql

COPY src/* $GOPATH/src/

CMD mysql start
WORKDIR $GOPATH
