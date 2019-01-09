FROM golang:1.9

# create project dir
RUN mkdir -p $GOPATH/src/github.com/maximzasorin/highloadcup-2

# move to project dir
WORKDIR $GOPATH/src/github.com/maximzasorin/highloadcup-2

# add source code
ADD ./ .

# build project
RUN go build -o highloadcup-2 ./src/*.go

CMD ["./highloadcup-2"]