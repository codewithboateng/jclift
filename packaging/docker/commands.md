# Local CI

make ci-local

# Build image + run analyzer inside container

make docker-build
make docker-run

# Compose (from packaging/docker)

cd packaging/docker
docker compose up --build --abort-on-container-exit
