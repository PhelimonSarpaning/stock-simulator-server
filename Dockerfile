FROM alpine:3.10.3
ADD stock-simulator-server /opt/server/
ADD config /opt/server/config
RUN chmod +x /opt/server/stock-simulator-server
EXPOSE 8000
CMD ["./opt/server/stock-simulator-server"]