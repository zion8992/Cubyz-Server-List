#!/usr/bin/env python3
import configparser
import os
import queue
import re
import sys
import threading
import time
from pathlib import Path
from typing import Callable, Dict, List, Tuple
import urllib.parse
import urllib.request
import urllib.error

Spark_server = None
user_api_token = None


def send_Spark_update(update: str) -> None:
    global Spark_server, user_api_token

    error_pattern = re.compile(r"^\[(?P<code>[^]]+)\]")

    url = f"{Spark_server}/api/v1/sparkUpdate"
    data = urllib.parse.urlencode(
        {"token": user_api_token, "update": update}
    ).encode("utf-8")
    req = urllib.request.Request(url, data=data, method="POST")
    try:
        urllib.request.urlopen(req, timeout=10)
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        print(f"HTTP error {e.code}: {body}", flush=True)

        m = error_pattern.match(body)
        code = m.group("code") if m else None

        match code:
            case "0":
                print("Server failed to parse form sent from this Spark client. Please ensure Spark or Server list are up-to-date. If the issue persists, please contact the server list's support.")
                sys.exit()
            case "1":
                print("Server couldn't check if your API token is valid. Try again, if the issue persists, please contact the server list's support.")
                sys.exit()
            case "2":
                print("Invalid API token.")
                sys.exit()
            case "3":
                print("Server couldn't get the server that corresponds to your API token. Try again, if the issue persists, please contact the server list's support.")
                sys.exit()
            case "4":
                print("Invalid data sent to server (specifically the update type). Try updating Spark.")
                sys.exit()
            case "5":
                print("Server couldn't update your server. Try again, if the issue persists, please contact the server list's support.")
                sys.exit()
            case _:
                print(f"Unknown error code in response: {body}")
        sys.exit()

    except Exception as e:
        print(f"Failed to send update: {e}", flush=True)
        m = error_pattern.match(str(e))
        code = m.group("code") if m else None
        match code:
            case "0" | "1" | "2" | "3" | "4" | "5":
                print(f"Error code {code} detected in exception message.")
            case _:
                print("No recognized error code in exception message.")
        


CONFIG_PATH = Path("sparkConfig.ini")
CUBYZ_VERSION = "0.2.0"

DEFAULTS: Dict[str, str] = {
    "user_api_token": "",
    "cubyz_log": "test-server/logs/latest.log",
    "Spark_server": "servers.ashframe.net",
}

PATTERNS: List[Tuple[str, re.Pattern, Callable[[re.Match], Dict]]] = [
    (
        "serverReady",
        re.compile(r"\[info\]: Finished world assets"),
        lambda m: {},
    ),
    (
        "join",
        re.compile(r"User (?P<user>.+?) joined using version (?P<version>.+)"),
        lambda m: m.groupdict(),
    ),
    (
        "leave",
        re.compile(r"Chat: (?P<user>.+?) left"),
        lambda m: m.groupdict(),
    ),
    (
        "death",
        re.compile(r"Chat: (?P<user>.+?) died of fall damage"),
        lambda m: m.groupdict(),
    ),
    (
        "lag",
        re.compile(r"\[warning\]: The server is lagging behind by (?P<lagtime>.+)"),
        lambda m: m.groupdict(),
    ),
]


def ensure_config(path: Path) -> configparser.ConfigParser:
    config = configparser.ConfigParser()
    if not path.exists():
        config["settings"] = DEFAULTS
        with open(path, "w", encoding="utf-8") as f:
            config.write(f)
        print(f"Created default config: {path}", file=sys.stderr)
    else:
        config.read(path)
    return config


def classify(line: str) -> Tuple[str, Dict]:
    for event, pattern, extractor in PATTERNS:
        match = pattern.search(line)
        if match:
            return event, extractor(match)
    return "default", {"raw": line}


def reader(log_path: str, q: queue.Queue, stop: threading.Event) -> None:
    path = Path(log_path)
    f = None
    try:
        while not stop.is_set():
            if not path.exists():
                time.sleep(0.5)
                continue

            if f is None:
                f = open(path, "r", encoding="utf-8", errors="ignore")
                f.seek(0, os.SEEK_END)

            line = f.readline()
            if not line:
                try:
                    size = path.stat().st_size
                    if f.tell() > size:
                        f.seek(0, os.SEEK_END)
                except OSError:
                    f.close()
                    f = None
                time.sleep(0.1)
                continue

            line = line.rstrip("\n")
            if line and not line.startswith("[debug]"):
                event, data = classify(line)
                q.put({"event": event, "data": data})
    finally:
        if f:
            f.close()


def printer(q: queue.Queue, stop: threading.Event) -> None:
    playercount = 0
    server_version = "0.3.0"
    latest_lag = ""

    while not stop.is_set() or not q.empty():
        try:
            item = q.get(timeout=0.1)
        except queue.Empty:
            continue

        event = item["event"]
        data = item["data"]

        match event:
            case "serverReady":
                print(f"[{event}] {data}", flush=True)
                print("Server is ready")
                send_Spark_update("serverReady")

            case "join":
                print(
                    f"[{event}] {data['user']} joined using {data['version']}",
                    flush=True,
                )
                playercount += 1
                server_version = data["version"]
                print(f"Player count: {playercount}")
                print(f"Set version: {server_version}")
                send_Spark_update("playerJoin")

            case "leave":
                print(f"[{event}] {data['user']} left", flush=True)
                playercount -= 1
                print(f"Player count: {playercount}")
                send_Spark_update("playerLeave")

            case "death":
                print(f"[{event}] {data['user']} died", flush=True)
                send_Spark_update("playerDeath")

            case "lag":
                print(
                    f"[{event}] The server is lagging behind by {data['lagtime']}"
                )
                latest_lag = data["lagtime"]
                send_Spark_update("serverLag")

            case _:
                print(f"[{event}] {data}", flush=True)


def main() -> None:
    global Spark_server, user_api_token

    print(f"You are running Spark for Cubyz version {CUBYZ_VERSION}.")
    print(f"There may be compatibility issues if your server isn't running {CUBYZ_VERSION}.")

    config = ensure_config(CONFIG_PATH)

    log_path = config.get(
        "settings", "cubyz_log", fallback=DEFAULTS["cubyz_log"]
    )
    user_api_token = config.get(
        "settings", "user_api_token", fallback=DEFAULTS["user_api_token"]
    )
    Spark_server = config.get(
        "settings", "Spark_server", fallback=DEFAULTS["Spark_server"]
    )

    print(f"Spark server: {Spark_server}")
    print(f"Spark cubyz log: {log_path}")
    print(f"Spark config file: {config_path}")

    if not Path(log_path).exists():
        print(
            f"Cannot open file: {log_path}. Please ensure the path is correct "
            f"in your {CONFIG_PATH}",
            file=sys.stderr,
        )
        sys.exit(1)

    q: queue.Queue = queue.Queue()
    stop = threading.Event()

    t_reader = threading.Thread(
        target=reader, args=(log_path, q, stop), daemon=True
    )
    t_printer = threading.Thread(
        target=printer, args=(q, stop), daemon=True
    )

    t_reader.start()
    t_printer.start()

    try:
        while True:
            time.sleep(0.5)
    except KeyboardInterrupt:
        stop.set()
        t_reader.join()
        t_printer.join()
        print("Shutting Spark down...")
        send_Spark_update("serverOff")


if __name__ == "__main__":
    main()
