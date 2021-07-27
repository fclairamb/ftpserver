# File system loading

On top of the supported file system and their associated config, 
they are few parameters that can be enabled.

## Read-only
This blocks any writing call to the underlying file system. It can be abled by the `read_only` boolean parameter.

```json
{
   "version": 1,
   "accesses": [
      {
         "read_only": true,
         
         // The usual FS config:
         "user": "test",
         "pass": "test",
         "fs": "os",
         "params": {
            "basePath": "/tmp"
         }
      }
   ]
}
``` 

## Shared
This makes sure the underlying filesystem is loaded once per use. It can be enabled by the `shared` boolean parameter.

```json
{
   "version": 1,
   "accesses": [
      {
         "shared": true,
        // The usual FS config:
        "user": "test",
         "pass": "test",
         "fs": "os",
         "params": {
            "basePath": "/tmp"
         }
      }
   ]
}
``` 

## Sync & Delete
This creates a temporary local folder that will be replicated to a destination file system.


```json
{
   "version": 1,
   "accesses": [
      {
        "sync_and_delete": {
            "enable": true,         // To enable it
            "directory": "/tmp/dir" // Temporary directory (optional)
         },
         // The usual FS config:
         "user": "test",
         "pass": "test",
         "fs": "os",
         "params": {
            "basePath": "/target"
         }
      }
   ]
}
``` 

## Logging
```json
{
   "version": 1,
   "accesses": [
      {
        "logging": {
            "file_accesses": true, // This logs every single file access
            "ftp_exchanges": true // This logs every single FTP command
        },
        // The usual FS config:
         "user": "test",
         "pass": "test",
         "fs": "os",
         "params": {
            "basePath": "/target"
         }
      }
   ]
}
``` 