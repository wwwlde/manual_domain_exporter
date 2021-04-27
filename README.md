# (Manual) Domain Exporter

Для доменов в зонах kz/tj/cy во whois базах нет даты истечения домена и нет удобного API чтобы его получить. Будем задавать ее вручную в файле `domains.json`.

Пример:

```bash
$ cat domains.json
{
  "domains": [
    {
      "name": "example.tj",
      "expire": "2024-01-03"
    },
    {
      "name": "example.kz",
      "expire": "2024-07-11"
    }
  ]
}
$ ./domain_exporter_static
level=info ts=2021-04-27T00:14:39.449Z caller=main.go:65 msg="Starting domain_exporter" version=0.0.1
level=info ts=2021-04-27T00:14:39.456Z caller=main.go:125 msg=Listening port=:9203
level=info ts=2021-04-27T00:14:39.456Z caller=main.go:137 domain:=example.tj days=978 date=2024-01-01T00:00:00Z
level=info ts=2021-04-27T00:14:39.456Z caller=main.go:137 domain:=example.kz days=1170 date=2024-07-11T00:00:00Z
$ curl -s localhost:9203/metrics | grep domain
# HELP domain_expiration Days until the WHOIS record states this domain will expire
# TYPE domain_expiration gauge
domain_expiration{domain="example.kz"} 1170
domain_expiration{domain="example.tj"} 980
# HELP domain_manual_expiration That the domain expiration date was set manualy
# TYPE domain_manual_expiration gauge
domain_manual_expiration{domain="example.kz"} 1
domain_manual_expiration{domain="example.tj"} 1
```
