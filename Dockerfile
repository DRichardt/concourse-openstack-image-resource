FROM scratch

ENV PATH /opt/resource:$PATH
COPY bin/ /opt/resource/
