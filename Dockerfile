# Should be started with:
# docker run -ti -p 2121-2200:2121-2200 ftpserver/ftpserver
FROM alpine:latest
EXPOSE 2121-2200
RUN mkdir -p /data
COPY sample/conf/settings.toml /etc/ftpserver.conf
COPY ftpserver /bin/ftpserver
ENTRYPOINT [ "/bin/ftpserver", "-conf=/etc/ftpserver.conf", "-data=/data" ]
