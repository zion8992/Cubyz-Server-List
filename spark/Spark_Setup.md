# Setup and Run Spark
Spark is a simple script that allows you to set an automated player count and server status on the list.
<br><br>

### Downloading
You can download `spark.py` from the [Ironite Github Releases Page](https://github.com/zion8992/ironite/releases).
<br>

### Configuring
FILE: `sparkConfig.ini`<br>

```ini
[settings]
user_api_token = <your api token>
cubyz_log = logs/latest.log
spark_server = https://servers.ashframe.net
```
<br>

**API Token**<br>
You can generate an API token for Spark through Ironite.<br>
**Account** -> **API Tokens** -> **Generate New Token**<br>
For the `Type` select `Spark` and for the server, choose the server you want to set the status.<br>
<br>

**Cubyz Log**<br>
Enter the location of Cubyz's Server `latest.log`. Make sure that when you run Cubyz, you see the log *live update*, if you don't, it is most likely the wrong file.<br>
<br>

**Spark Server**<br>
Default is [https://servers.ashframe.net](https://servers.ashframe.net). Change this if you are using a different instance of Ironite.<br>
<br>

## Runnning
Requires `python3.14` or later.<br>

```sh
python3 spark.py
```