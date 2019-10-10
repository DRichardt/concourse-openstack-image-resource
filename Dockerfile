FROM ubuntu

ENV PATH /opt/resource:$PATH
RUN apt-get update && apt-get install -y musl-dev && ln -s /usr/lib/x86_64-linux-musl/libc.so /lib/libc.musl-x86_64.so.1
COPY bin/ /opt/resource/
