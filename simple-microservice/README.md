# operator-sample-go: simple-microservice

Work in progress ...

## Usage

mvn clean quarkus:dev

http://localhost:8081/hello

mvn clean install

podman build -f src/main/docker/Dockerfile.jvm -t nheidloff/simple-microservice .

podman run -i --rm -p 8081:8081 -e GREETING_MESSAGE=World nheidloff/simple-microservice

podman tag localhost/nheidloff/simple-microservice docker.io/nheidloff/simple-microservice

podman push docker.io/nheidloff/simple-microservice
