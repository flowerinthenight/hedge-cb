FROM golang:1.25.3-trixie
COPY . /go/src/github.com/flowerinthenight/hedge-cb/
WORKDIR /go/src/github.com/flowerinthenight/hedge-cb/example/
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates xz-utils && rm -rf /var/lib/apt/lists/* && \
    curl https://sh.rustup.rs -sSf | bash -s -- -y
RUN . "$HOME/.cargo/env"; ROOTDIR=$PWD; cd /tmp/ && git clone --single-branch -b main-2.0 https://github.com/aws/clock-bound && \
    cp -v clock-bound/clock-bound-ffi/include/clockbound.h /usr/include/ && \
    cd /tmp/clock-bound/clock-bound-ffi/ && \
    rustup target add x86_64-unknown-linux-musl && \
    cargo build --release --target=x86_64-unknown-linux-musl && \
    cd /tmp/ && wget https://musl.libc.org/releases/musl-1.2.5.tar.gz && \
    tar xvzf musl-1.2.5.tar.gz && cd musl-1.2.5/ && ./configure && make -s && make install && \
    cp -v /tmp/clock-bound/target/x86_64-unknown-linux-musl/release/libclockbound.a /usr/local/musl/lib/ && \
    cd /tmp/ && wget https://ziglang.org/download/0.15.2/zig-x86_64-linux-0.15.2.tar.xz && ls -laF && \
    tar -xvf zig-x86_64-linux-0.15.2.tar.xz && cd $ROOTDIR/ && \
    CC="/tmp/zig-x86_64-linux-0.15.2/zig cc -target x86_64-linux-musl -I/tmp/clock-bound/clock-bound-ffi/include -L/usr/local/musl/lib -lunwind" GOOS=linux GOARCH=amd64 \
    go build -v --ldflags '-linkmode=external -extldflags=-static'

FROM ubuntu:24.04
WORKDIR /app/
COPY --from=0 /go/src/github.com/flowerinthenight/hedge-cb/example/example .
ENTRYPOINT ["/app/example"]
CMD ["-db", "put-dsn-here"]
