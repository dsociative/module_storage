FROM debian

VOLUME [ "/data" ]
EXPOSE 8080

ADD ./module_storage /usr/local/bin/module_storage
ENTRYPOINT [ "/usr/local/bin/module_storage", "--path", "/data"]