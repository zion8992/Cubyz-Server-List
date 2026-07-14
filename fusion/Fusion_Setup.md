# Setup and Run Fusion
Fusion is a simple app that allows you to manage, create, edit and delete announcements for your server.
<br><br>

## Downloading
You can download `fusion.py` from the [Ironite Github Releases Page](https://github.com/zion8992/ironite/releases).
<br>

## Runnning

#### Requirements

- `python3.14` or later.
- `tkinter` python library
- `markdown` python library
- `tkinterweb` python library

Requirements (excluding `python`) can be installed with:
```sh
pip install tkinter markdown tkinterweb
```

<br><br>

#### Running

```sh
python3 fusion.py
```

## Configuring
FILE: `fusion.ini`<br>
The configuration file is managed by Fusion. You can change the configuration settings inside of Fusion.

```ini
[Settings]
theme = Midnight
mode = dark
show_intro = False
```
<br>

**Theme**<br>
Set the theme for your app. Themes can be found in the settings page of Fusion.
<br>

**Mode**<br>
Sets the theme variant. Can be `light` or `dark`.
<br>

**Show Intro**<br>
Wether or not to show the "Fusion" intro when the app starts.
<br>