# paradice_ftp
paradice_ftp is a powerful, 100% native (golang) ftp server that is production ready.

It can handle 1000's of connections and 1000's of files flying back and forward sideways under and through. It does not run out of file descriptions. It does not forget to close any socket connection or socket listener. Ahem hem, cough cough, looking at you https://github.com/goftp/server.

Enjoy.

FYI FTP is a big protocol and I only implemented the stuff I needed. Stuff that's here:

1) passive socket connections (not active ones)
2) uploading files (not downloading)
3) directory listing
4) user authentication (soon to suppport Bitium API)
5) Both EPSV and PASV commands
