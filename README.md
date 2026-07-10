## Running
To run Ironite, you can either use `run.sh` or a [prebuilt executable from the releases page](https://github.com/zion8992/ironite/releases).

Before running Ironite, you need to setup a MySQL database. You can use the existing scripts in the project source to run in docker or you can [set a MySQL database up yourself](https://dev.mysql.com/downloads/mysql/).

### Create/Run database in Docker
1. Clone the repository
2. Run `./setup.sh`
3. Run `mysqlpass="YOUR MYSQL PASSWORD" ./scripts/connect_db.sh" 
4. Paste the contents of `database_setup.sql` (located in the project root) into the MySQL console.


#### Run using Prebuilt Binary
Download the Ironite executable for your platform from the [downloads page](https://github.com/zion8992/ironite/releases). Once downloaded, open your terminal and run:

**Linux**
```sh
chmod +x ironite-linux-arch # replace with your architecture
./ironite-linux-arch # replace with your architecture
```

**Windows**
```sh
.\ironite-windows-arch # replace with your architecture
```

 If you changed the root database password, see the **Command Line Flags** below to configure Ironite to support your change. If you used the default, don't change anything.

#### Run using `run.sh`
Make sure you have [Go](https://go.dev) `1.25` or later installed. That's all Ironite needs.

```sh
cd ../ # cd out of scripts if you were in scripts/
go mod download
./run.sh
```

 If you changed the root database password, see the **Command Line Flags** below to configure Ironite to support your change. If you used the default, don't change anything.

#### Command Line Arguments
`dbpass`: Password for the root account of your mysql database, Default: `H0EeLfLnO,xDEVELOPERSx4c!#%`. <br>
`dburl`: URL of the database, Default: `127.0.0.1:3306`.<br>
`port`: Port to host http server, Default: `:8000`. <br>
#### Example Usage
```
./ironite-linux-amd64 -dbpass "my password" -dburl "my url"
```

### Spark
Spark is a python client for server admins to set server status on the server list. **Not required to run the server list**.
Instructions to setup and run Spark can be found [here](https://github.com/zion8992/ironite/blob/main/spark/Spark_Setup.md)

### Fusion
Fusion is GUI app that allows server admins to create announcements for their servers. **Not required for ironite to run**.
Instructions to setup and run Fusion can be found [here](https://github.com/zion8992/ironite/blob/main/fusion/Fusion_Setup.md)