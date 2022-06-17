# dockerlog2web

View your docker container logs on a simple website.
Ideal for exposing the logs of applications on test servers or other non-critical services.

![Screenshot](https://raw.githubusercontent.com/joschahenningsen/dockerlog2web/main/screenshot.png)

## Usage: 

`docker run -e CONTAINER=my-container -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock dockerlog2web`