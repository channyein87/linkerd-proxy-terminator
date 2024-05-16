FROM debian
COPY ./proxy-terminator /proxy-terminator
CMD ["/proxy-terminator"]