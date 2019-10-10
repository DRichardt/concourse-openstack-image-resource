FROM alpine

ENV PATH /opt/resource:$PATH
RUN apk add --no-cache ca-certificates libc6-compat && ln -s /lib/libpthread.so.0 /lib64/libpthread.so.2 && ln -s /lib/libpthread.so.0 /lib64/libpthread.so.0

COPY bin/ /opt/resource/
