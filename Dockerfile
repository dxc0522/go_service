FROM china-devops-docker-local.arf.tesla.cn/base-images/golang:1.18.3-alpine3.16 AS compiler

ENV GO111MODULE=on
ENV GOPROXY=http://goproxy-china-it.mo.tesla.cn:8082,https://goproxy.cn/,https://goproxy.io/,direct

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add build-base
WORKDIR /bjm

ADD . .

ARG MODULE_NAME
WORKDIR /bjm/internal/$MODULE_NAME

RUN GOOS=linux GOARCH=amd64 go build -tags musl -ldflags "-s -w" -o $MODULE_NAME.bin ./cmd/*
RUN mkdir -p dist && \
    cp $MODULE_NAME.bin dist && \
    if [ -d db ]; then cp -r db dist; fi

FROM china-devops-docker-local.arf.tesla.cn/base-images/alpine:3.16.0

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add ca-certificates tzdata

ARG MODULE_NAME
WORKDIR /bjm/$MODULE_NAME
RUN true
COPY --from=compiler /bjm/internal/$MODULE_NAME/dist .
RUN true

ARG MODULE_VERSION
ARG GIT_COMMIT
ARG GIT_BRANCH
RUN echo "$MODULE_VERSION" > version.txt
RUN echo "$GIT_COMMIT" > commit.txt
RUN echo "$(date "+%Y-%m-%d %H:%M:%S" )" > buildDate.txt
RUN echo "$GIT_BRANCH" > branch.txt

ENV MODULE_NAME=$MODULE_NAME

ENTRYPOINT /bjm/$MODULE_NAME/$MODULE_NAME.bin
