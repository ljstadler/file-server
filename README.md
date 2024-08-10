<div align=center>

# file-server

## HTTP/TUS file server

</div>

## Usage

### Docker Compose

```yaml
services:
    file-server:
        container_name: file-server
        environment:
            - MANAGE_TOKEN=${MANAGE_TOKEN}
            - VIEW_TOKEN=${VIEW_TOKEN}
        image: ghcr.io/ljstadler/file-server:latest
        ports:
            - 1323:1323
        volumes:
            - ./files:/file-server/files
```

### Docker Run

```bash
docker run -d -e MANAGE_TOKEN="{MANAGE_TOKEN}" -e VIEW_TOKEN="{VIEW_TOKEN}" --name file-server -p 1323:1323 -v ./files:/file-server/files ghcr.io/ljstadler/file-server
```

### Endpoints

-   Go to `{HOST}:{PORT}?auth={AUTH}` using either the `MANAGE_TOKEN` or `VIEW_TOKEN`
-   Do a `GET` request to the same endpoint with the `Accept` header set to `application/json` to get a JSON response
-   Do a `GET` request to `{HOST}:{PORT}/file/{NAME}?auth={AUTH}` to download a file
-   Do a `DELETE` request to the same endpoint to delete a file

## Screenshots

### Manage

![](./manage.png)

### View

![](./view.png)
