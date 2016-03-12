# paradise_ftp
paradise_ftp is a powerful, 100% native (golang) ftp server that is production ready.

It can handle 1000's of connections and 1000's of files flying back and forward sideways under and through. It does not run out of file descriptions. It does not forget to close any socket connection or socket listener. Ahem hem, cough cough, looking at you https://github.com/goftp/server.

Enjoy.

FYI FTP is a big protocol and I only implemented the stuff I needed. Stuff that's here:

 * passive socket connections (not active ones)
 * uploading files (not downloading)
 * directory listing
 * user authentication (soon to suppport Bitium API https://developer.bitium.com)
 * Both EPSV and PASV commands
 * uploads large files without reading entire file into memory
 * just simple if command == "this" else if command == "that" logic
 * uploads read first 512 bytes of file first into bufffer to check mime type

Sample Run:

```
$ ftp ftp://auser:secret@127.0.0.1:2121
Connected to 127.0.0.1.
220 Welcome to Paradise
331 User name ok, password required
230 Password ok, continue
Remote system type is UNIX.
Using binary mode to transfer files.
200 Type set to binary

ftp> dir
229 Entering Extended Passive Mode (|||55729|)
150 Opening ASCII mode data connection for file list
-rw-r--r-- 1 paradise ftp        13984 Mar 12 11:51 paradise.txt
-rw-r--r-- 1 paradise ftp        13984 Mar 12 11:51 paradise.txt
-rw-r--r-- 1 paradise ftp        13984 Mar 12 11:51 paradise.txt
-rw-r--r-- 1 paradise ftp        13984 Mar 12 11:51 paradise.txt
-rw-r--r-- 1 paradise ftp        13984 Mar 12 11:51 paradise.txt

226 Closing data connection, sent bytes
ftp> put file_driver.go 
local: file_driver.go remote: file_driver.go
229 Entering Extended Passive Mode (|||55732|)
150 Data transfer starting
100% |**********************************************************************|  4624        8.89 MiB/s    00:00 ETA
226 OK, received some bytes
4624 bytes sent in 00:00 (981.44 KiB/s)
ftp> 

```

Server Output:

```
$ ./paradise 
listening on:  localhost:2121
Got client on:  127.0.0.1:55728
127.0.0.1:55728 USER auser
127.0.0.1:55728 EPSV 
127.0.0.1:55728 LIST 
127.0.0.1:55728 EPSV 
127.0.0.1:55728 STOR file_driver.go
```
