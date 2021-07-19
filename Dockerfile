FROM ubuntu:18.04
MAINTAINER vinhph2
RUN apt-get update
COPY sidecar /bin/sidecar
ENTRYPOINT ["sidecar","-d"]
#RUN go build -o /bin/sidecar
