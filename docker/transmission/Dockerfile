FROM centos:7.9.2009 AS builder

ARG LIB_EVENT_VER=2.1.11
ARG TRANSMISSION_VER=3.00

# dependencies
RUN yum -y install gcc make openssl-devel curl-devel intltool gcc-c++ m4 automake libtool gettext wget


RUN cd /usr/src && \
    wget https://github.com/libevent/libevent/releases/download/release-$LIB_EVENT_VER-stable/libevent-$LIB_EVENT_VER-stable.tar.gz && \
    tar zxvf libevent-$LIB_EVENT_VER-stable.tar.gz && \
    cd libevent-$LIB_EVENT_VER-stable && \
    ./configure --prefix=/opt/libevent && \
    make && \
    make install

RUN cd /usr/src && \
    wget https://github.com/transmission/transmission/releases/download/$TRANSMISSION_VER/transmission-$TRANSMISSION_VER.tar.xz && \
    tar xvf transmission-$TRANSMISSION_VER.tar.xz && \
    cd transmission-$TRANSMISSION_VER && \
    PKG_CONFIG_PATH=/opt/libevent/lib/pkgconfig ./configure --prefix=/opt/transmission && \
    make && \
    make install

FROM centos:7.9.2009 AS release
COPY --from=builder /opt /opt
ENTRYPOINT ["/opt/transmission/bin/transmission-daemon", "-f"]
