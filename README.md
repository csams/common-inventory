# Common Inventory
This repository explores designs for a common inventory system.  The models and even behavior in it are
currently wrong, but it provides a place to start and sketches ideas.

Run `make build` and then `./bin/common-inventory help`.

Also try `./bin/comment-inventory serve help`

Completion is available:

```
/bin/common-inventory completion bash | source
```

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
curl -H "Content-Type: application/json" -d '{"DisplayName": "Example Cluster" "ResourceType": "k8s-cluster", "ResourceIdAlias": "1", "Data": {"ApiServer": "www.example.com/api-server"}}}' http://localhost:9080/api/inventory/v1.0/resources
```

Then run

```bash
curl '127.0.0.1:9080/api/v1.0/resources' | jq .
```
