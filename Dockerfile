FROM ubuntu:18.04
MAINTAINER vngcloud
RUN apt-get update
COPY sidecar /bin/sidecar
ENTRYPOINT ["sidecar","-d"]
#RUN go build -o /bin/sidecar
