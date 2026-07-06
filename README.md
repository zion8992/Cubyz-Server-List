## Ironite
Open source server list for Cubyz made using Go.

### Run

#### Prebuilt Binaries
Prebuilt executables can be found [here](https://github.com/zion8992/ironite/releases)

#### Using Script
Requires Go `1.25` or later to run. You can download Go from [go.dev](https://go.dev).
Runs on all platforms.

```bash
./run.sh <arguments>
```

#### Command Line Arugments
`dbpass`: Password for the root account of your mysql database. Latest release of ironite auto assumes you are running MySQL on `127.0.0.1:3306`, Default: `H0EeLfLnO,xDEVELOPERSx4c!#%`.<br>
`port`: Port to host http server, Default: `:8000`.

### Build

```bash
GOOS=linux GOARCH=amd64 go build -o ironite ./src/ # for linux amd64
GOOS=windows GOARCH=amd64 go build -o ironite.exe ./src/ # for windows amd64
GOOS=darwin GOARCH=amd64 go build -o ironite ./src/ # for macOS amd64
```
<br>

### Spark
Instructions to setup and run Spark can be found [here](https://github.com/zion8992/ironite/blob/main/spark/Spark_Setup.md)

### Fusion
Instructions to setup and run Fusion can be found [here](https://github.com/zion8992/ironite/blob/main/fusion/Fusion_Setup.m)