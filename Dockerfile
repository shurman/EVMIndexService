FROM ubuntu:latest

RUN apt-get update && apt-get install -y wget git gcc vim software-properties-common mysql-server
RUN add-apt-repository -y ppa:ethereum/ethereum && apt-get update && apt-get install -y ethereum
RUN wget -P /tmp https://go.dev/dl/go1.18.3.linux-amd64.tar.gz &&  tar -C /usr/local -xzf /tmp/go1.18.3.linux-amd64.tar.gz && rm /tmp/go1.18.3.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN cd $GOPATH/src && go mod init ethmod && go get -u github.com/ethereum/go-ethereum gorm.io/gorm gorm.io/driver/mysql

COPY mysqlsetup.sh /tmp
RUN sh /tmp/mysqlsetup.sh

WORKDIR $GOPATH
