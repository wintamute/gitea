#!/usr/bin/env bash

VERSION=1.27.0-cssa
docker build --build-arg GITEA_VERSION=$VERSION -t registry.dev.cssa.de/library/gitea:$VERSION  .
docker push registry.dev.cssa.de/library/gitea:$VERSION
