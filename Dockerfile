FROM golang:1.22 as build

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 编译后端

WORKDIR /go/cache

ADD go.mod .
ADD go.sum .
RUN go mod download

WORKDIR /go/release
ADD . .


RUN CGO_ENABLED=0 GOOS=linux  go build -o wk ./main.go

FROM alpine as prod
WORKDIR /home
COPY --from=build /go/release/wk /home
# ENTRYPOINT ["sh","-c","/home/wk context add demo --server ${WK_SERVER} --token ${WK_TOKEN} & tail -f /dev/null"]
