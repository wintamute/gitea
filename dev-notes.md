# Dev Notes

# Integration tests


Make sure to perform a clean build before running tests:
```shell
make clean build
```

Integration tests can be run with make commands for the
appropriate backends, namely:
```shell
make test-sqlite
```

## DB

```shell
docker exec -it gitea_postgres_16 bash
```

```shell
psql -U $POSTGRES_USER -d $POSTGRES_DB
```
```shell
export VERSION=1.27.0-cssa
docker build --build-arg GITEA_VERSION=$VERSION -t registry.dev.cssa.de/library/gitea:$VERSION  .
docker push registry.dev.cssa.de/library/gitea:$VERSION
```


