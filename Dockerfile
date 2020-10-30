# Should be started with:
# docker run -ti -p 2121-2130:2121-2130 ftpserver/ftpserver
FROM alpine:3.12.1
EXPOSE 2121-2130
COPY ftpserver /bin/ftpserver
ENTRYPOINT [ "/bin/ftpserver" ]
