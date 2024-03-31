FROM golang:1.19 AS build-env  
ENV GOPROXY https://goproxy.cn  
ADD . /go/src/app  
WORKDIR /go/src/app  
RUN go mod tidy  
RUN GOOS=linux GOARCH=386 go build -v -o /go/src/app/go-interface-health-check  
  
FROM alpine  
COPY --from=build-env /go/src/app/go-interface-health-check /usr/local/bin/go-interface-health-check  
COPY --from=build-env /go/src/app/config.yaml /opt/  
WORKDIR /opt  
EXPOSE 8080 
CMD [ "go-interface-health-check","--config=/opt/config.yaml" ]