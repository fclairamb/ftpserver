
You should have a config file looking like this one:
```json
{
  "accesses": [
   {
      "fs": "keycloak",
      "params": {
        "keycloak_client_id": "*****",
        "keycloak_client_secret": "*****",
        "keycloak_url": "http://localhost:8080",
        "keycloak_realm": "master",
        "base_path": "/ftp"
      }
    }
  ]
}
```
