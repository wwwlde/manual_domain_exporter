FROM scratch
MAINTAINER Denis Lemeshko <lde@linuxhelp.com.ua>
COPY bin/domain_exporter_static /domain_exporter
ENTRYPOINT ["/domain_exporter"]
