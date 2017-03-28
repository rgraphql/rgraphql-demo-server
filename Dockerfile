FROM alpine:edge
ADD ./rgraphql-demo-server /
RUN chmod +x /rgraphql-demo-server
EXPOSE 3001
CMD ["/rgraphql-demo-server"]
