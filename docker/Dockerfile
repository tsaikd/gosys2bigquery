FROM alpine:3.6

ENV GOSYS2BIGQUERY_VERSION 0.0.1

ADD https://github.com/tsaikd/gosys2bigquery/releases/download/${GOSYS2BIGQUERY_VERSION}/gosys2bigquery-Linux-x86_64 /usr/local/bin/gosys2bigquery

RUN chmod +x /usr/local/bin/gosys2bigquery

CMD ["gosys2bigquery"]
