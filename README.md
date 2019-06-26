# Pharos

🚨**Warning**: this project is currently under active development and is not considered stable or
functional. 🚨

## Overview
Pharos is an open-source Kubernetes cluster discovery and configuration distribution tool designed
to work nicely with [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator).


## Testing Locally
Build the Pharos API server and Pharos CLI:
```
make install
make build
```

Set up the local database.
```
make setup
```

You can connect to the local database using the following command:
```bash
psql -U pharos_admin -d pharos
```

You can insert as many test clusters as you like using psql. The values in order are: cluster name,
cluster environment, cluster server URL, cluster authority data, cluster deletion status, and
cluster active status.
```sql
psql (9.6.10)
Type "help" for help.

pharos=> INSERT INTO clusters VALUES ('test-111111', 'test', 'https://test.com', 'test', false, true);
```

Start the Pharos server:
```bash
make start
```

From within the pharos directory, you can run the Pharos CLI with the config flag `-c` to specify a
specific config file. You can use the config file in the testdata folder to connect automatically
to your local Pharos server.

Example: `bin/pharos clusters list -c pkg/pharos/testdata/pharosConfig`

**IMPORTANT NOTE: If you're running a command that will edit or create a new kubeconfig, you will
need to run it with the file flag `-f` to prevent overwriting or modifying your existing kubeconfig
file at `$HOME/.kube/config`.** Some commands that edit kubeconfig files will also create new ones,
so for those commands you can specify any file, even ones that don't exist. The only commands that
require an existing kubeconfig file are `pharos clusters switch` and `pharos clusters current`.

Example: `bin/pharos clusters get sandbox -c pkg/pharos/testdata/pharosConfig -f test/test1` This
command will create a new kubeconfig file at `./test/test1` if it succeeds.

Happy testing!
