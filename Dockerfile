# Should be started with:
# docker run -ti -p 2121-2200:2121-2200 ftpserver/ftpserver
FROM alpine:3.11.6
EXPOSE 2121-2200
RUN mkdir -p /data
COPY ftpserver /bin/ftpserver
ENTRYPOINT [ "/bin/ftpserver", "-data=/data" ]
