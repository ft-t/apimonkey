FROM golang:1.22-bookworm
RUN apt-get update && apt-get install zip git -y
ADD . /src
WORKDIR /src
RUN GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o com.ftt.apimonkey.exe
RUN mkdir -p /dist/com.ftt.apimonkey.sdPlugin
RUN cp -r ./resources/* /dist/com.ftt.apimonkey.sdPlugin/
RUN cp com.ftt.apimonkey.exe /dist/com.ftt.apimonkey.sdPlugin/
WORKDIR /dist
RUN zip -r com.ftt.apimonkey.sdPlugin.zip com.ftt.apimonkey.sdPlugin