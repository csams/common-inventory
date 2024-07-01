# Common Inventory

This repository explores designs for a common inventory system.

Run `make build` and then `./bin/common-inventory help`.

You can set options on the command line, with [environment variables](https://pkg.go.dev/github.com/spf13/viper@v1.18.2#AutomaticEnv), or with a `.common-inventory.yaml` configuration file.

```yaml
server:
  address: localhost:9080
storage:
  database: sqlite3
  sqlite3:
    dsn: inventory.db
  postgres:
    host: 'localhost'
    port: '5432'
    dbname: inventory
```

All command line options can be given in the config file.

`storage.database` selects the database configuration to use.

First run `./bin/common-inventory migrate`
Then run `./bin/common-inventory serve`

In a separate terminal, run

```bash
❯ curl -H "Content-Type: application/json" -d '{"Metadata": {"DisplayName": "Example Host", "Reporter": "robot"}, "Fqdn": "www.example.com"}' http://localhost:9080/api/v1.0/hosts/
❯ curl -H "Content-Type: application/json" -d '{"Metadata": {"DisplayName": "Example Cluster", "Reporter": "robot"}, "ApiServer": "my.k8s.cluster.com"}' http://localhost:9080/api/v1.0/clusters/
```

Then run

```bash
❯ curl '127.0.0.1:9080/api/v1.0/hosts' | jq .
❯ curl '127.0.0.1:9080/api/v1.0/clusters' | jq .
❯ curl '127.0.0.1:9080/api/v1.0/resources' | jq .
```
