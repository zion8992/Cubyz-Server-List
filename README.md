## Ironite
Open source server list for Cubyz made using Go.

### Run
Requires Go `1.25` or later to run. You can download Go from (https://go.dev)[go.dev].
Runs on all platforms.

```bash
./run.sh
```

### Build

```bash
GOOS=linux GOARCH=amd64 go build -o ironite ./src/ # for linux
GOOS=windows GOARCH=amd64 go build -o ironite.exe ./src/ # for windows
GOOS=darwin GOARCH=amd64 go build -o ironite ./src/ # for macOS
```
<br>

**NOTE**: Final executable requires the `templates` directory from the source code.