import tkinter as tk
from tkinter import ttk, messagebox, simpledialog, filedialog
import markdown
import xml.etree.ElementTree as ET
import configparser
import os
from datetime import datetime  # >>> added

try:
    from tkinterweb import HtmlFrame
except ImportError:
    HtmlFrame = None

CONFIG_FILE = "fusion.ini"
INTRO_DURATION_MS = 2000
INTRO_FADE_STEPS = 27

# >>> Go uses a reference-date layout ("Mon Jan 2 15:04:05 MST 2006") instead
#     of specifiers like yyyy-mm-dd. Each token is a fixed component of that
#     reference time (2006=year, 01=month, 02=day, 15=hour, 04=min, 05=sec,
#     MST=tz abbrev). We replicate that exact layout in Python's strftime.
GO_TIME_LAYOUT_ABBR = "%b %d %H:%M:%S %Z %Y"   # Mon Jan 2 15:04:05 MST 2006
GO_TIME_LAYOUT_OFFSET = "%b %d %H:%M:%S %z %Y" # Mon Jan 2 15:04:05 -0700 2006


def go_format_now():
    """Return the current datetime as a string in Go's time layout format."""
    now = datetime.now().astimezone()
    tz_abbr = now.tzname() or ""
    # Prefer the timezone abbreviation ("MST") like Go's reference layout;
    # fall back to the numeric offset ("-0700") when no abbreviation exists.
    if tz_abbr and not tz_abbr.startswith(("+", "-")):
        return now.strftime(GO_TIME_LAYOUT_ABBR)
    return now.strftime(GO_TIME_LAYOUT_OFFSET)


class AnnouncementApp:
    THEMES = {
    "Midnight": {
        "dark": {
            "bg": "#1e1e2e",
            "sidebar": "#252537",
            "accent": "#7aa2f7",
            "accent_hover": "#bb9af7",
            "text": "#c0caf5",
            "muted": "#565f89",
            "card": "#2a2b3d",
            "selected": "#3b3d56",
            "editor_bg": "#1a1b26",
            "editor_fg": "#c0caf5",
        },
        "light": {
            "bg": "#f5f5f7",
            "sidebar": "#eaeaef",
            "accent": "#3b82f6",
            "accent_hover": "#2563eb",
            "text": "#1e1e2e",
            "muted": "#6b7280",
            "card": "#ffffff",
            "selected": "#d1d5db",
            "editor_bg": "#ffffff",
            "editor_fg": "#1e1e2e",
        },
    },
    "Ocean": {
        "dark": {
            "bg": "#0f172a",
            "sidebar": "#1e293b",
            "accent": "#38bdf8",
            "accent_hover": "#7dd3fc",
            "text": "#e2e8f0",
            "muted": "#64748b",
            "card": "#334155",
            "selected": "#475569",
            "editor_bg": "#1e293b",
            "editor_fg": "#e2e8f0",
        },
        "light": {
            "bg": "#f0f9ff",
            "sidebar": "#e0f2fe",
            "accent": "#0284c7",
            "accent_hover": "#0369a1",
            "text": "#0f172a",
            "muted": "#475569",
            "card": "#ffffff",
            "selected": "#bae6fd",
            "editor_bg": "#ffffff",
            "editor_fg": "#0f172a",
        },
    },
    "Forest": {
        "dark": {
            "bg": "#1a1c18",
            "sidebar": "#252a21",
            "accent": "#84cc16",
            "accent_hover": "#a3e635",
            "text": "#e8e8e6",
            "muted": "#6b7280",
            "card": "#2f3529",
            "selected": "#3f4637",
            "editor_bg": "#1f241c",
            "editor_fg": "#e8e8e6",
        },
        "light": {
            "bg": "#f4f7f0",
            "sidebar": "#e8ede4",
            "accent": "#4d7c0f",
            "accent_hover": "#3f6212",
            "text": "#1a1c18",
            "muted": "#4b5563",
            "card": "#ffffff",
            "selected": "#d9e6cc",
            "editor_bg": "#ffffff",
            "editor_fg": "#1a1c18",
        },
    },
    "Sunset": {
        "dark": {
            "bg": "#2a1b1b",
            "sidebar": "#3d2525",
            "accent": "#fb7185",
            "accent_hover": "#fda4af",
            "text": "#f5e6e6",
            "muted": "#9ca3af",
            "card": "#4a2c2c",
            "selected": "#5c3a3a",
            "editor_bg": "#331f1f",
            "editor_fg": "#f5e6e6",
        },
        "light": {
            "bg": "#fff5f5",
            "sidebar": "#ffe4e4",
            "accent": "#e11d48",
            "accent_hover": "#be123c",
            "text": "#2a1b1b",
            "muted": "#6b7280",
            "card": "#ffffff",
            "selected": "#fecdd3",
            "editor_bg": "#ffffff",
            "editor_fg": "#2a1b1b",
        },
    },
    "Sakura": {
        "dark": {
            "bg": "#2a1f24",
            "sidebar": "#3d2d35",
            "accent": "#f9a8d4",
            "accent_hover": "#f472b6",
            "text": "#fce7f3",
            "muted": "#a78b96",
            "card": "#4a2f3a",
            "selected": "#5c3a48",
            "editor_bg": "#33222a",
            "editor_fg": "#fce7f3",
        },
        "light": {
            "bg": "#fdf2f8",
            "sidebar": "#fce7f3",
            "accent": "#db2777",
            "accent_hover": "#be185d",
            "text": "#31262b",
            "muted": "#8b6e7a",
            "card": "#ffffff",
            "selected": "#fbcfe8",
            "editor_bg": "#ffffff",
            "editor_fg": "#31262b",
        },
    },
    "Lavender": {
        "dark": {
            "bg": "#1e1b2e",
            "sidebar": "#29233d",
            "accent": "#c084fc",
            "accent_hover": "#d8b4fe",
            "text": "#ede9fe",
            "muted": "#8b7aa0",
            "card": "#352d4d",
            "selected": "#453a63",
            "editor_bg": "#221d33",
            "editor_fg": "#ede9fe",
        },
        "light": {
            "bg": "#f7f5ff",
            "sidebar": "#ede9fe",
            "accent": "#7c3aed",
            "accent_hover": "#6d28d9",
            "text": "#221d33",
            "muted": "#6b5b7f",
            "card": "#ffffff",
            "selected": "#ddd6fe",
            "editor_bg": "#ffffff",
            "editor_fg": "#221d33",
        },
    },
    "Coffee": {
        "dark": {
            "bg": "#231c18",
            "sidebar": "#322622",
            "accent": "#d97706",
            "accent_hover": "#f59e0b",
            "text": "#efe6dd",
            "muted": "#9c8b7e",
            "card": "#3d2f28",
            "selected": "#4f3d35",
            "editor_bg": "#2a211c",
            "editor_fg": "#efe6dd",
        },
        "light": {
            "bg": "#f7f3ef",
            "sidebar": "#efe6dd",
            "accent": "#92400e",
            "accent_hover": "#78350f",
            "text": "#231c18",
            "muted": "#7c6a5d",
            "card": "#ffffff",
            "selected": "#e7d5c4",
            "editor_bg": "#ffffff",
            "editor_fg": "#231c18",
        },
    },
    "Cyberpunk": {
        "dark": {
            "bg": "#0a0a0f",
            "sidebar": "#14141f",
            "accent": "#f0abfc",
            "accent_hover": "#e879f9",
            "text": "#f0f0ff",
            "muted": "#6b7280",
            "card": "#1a1a2e",
            "selected": "#2d2d4a",
            "editor_bg": "#0f0f1a",
            "editor_fg": "#f0f0ff",
        },
        "light": {
            "bg": "#f8f8ff",
            "sidebar": "#e6e6fa",
            "accent": "#d946ef",
            "accent_hover": "#c026d3",
            "text": "#0a0a0f",
            "muted": "#6b7280",
            "card": "#ffffff",
            "selected": "#f5d0fe",
            "editor_bg": "#ffffff",
            "editor_fg": "#0a0a0f",
        },
    },
    "Solarized": {
        "dark": {
            "bg": "#002b36",
            "sidebar": "#073642",
            "accent": "#2aa198",
            "accent_hover": "#268bd2",
            "text": "#eee8d5",
            "muted": "#657b83",
            "card": "#073642",
            "selected": "#586e75",
            "editor_bg": "#00212b",
            "editor_fg": "#eee8d5",
        },
        "light": {
            "bg": "#fdf6e3",
            "sidebar": "#eee8d5",
            "accent": "#268bd2",
            "accent_hover": "#2aa198",
            "text": "#073642",
            "muted": "#586e75",
            "card": "#ffffff",
            "selected": "#d1cbb8",
            "editor_bg": "#ffffff",
            "editor_fg": "#073642",
        },
    },
    "Dracula": {
        "dark": {
            "bg": "#282a36",
            "sidebar": "#313442",
            "accent": "#ff79c6",
            "accent_hover": "#bd93f9",
            "text": "#f8f8f2",
            "muted": "#6272a4",
            "card": "#3a3c4e",
            "selected": "#44475a",
            "editor_bg": "#22232e",
            "editor_fg": "#f8f8f2",
        },
        "light": {
            "bg": "#f8f8f2",
            "sidebar": "#e8e8e3",
            "accent": "#d63384",
            "accent_hover": "#8b5cf6",
            "text": "#282a36",
            "muted": "#6272a4",
            "card": "#ffffff",
            "selected": "#c9c9c3",
            "editor_bg": "#ffffff",
            "editor_fg": "#282a36",
        },
    },
    "Monokai": {
        "dark": {
            "bg": "#272822",
            "sidebar": "#32332e",
            "accent": "#a6e22e",
            "accent_hover": "#66d9ef",
            "text": "#f8f8f2",
            "muted": "#75715e",
            "card": "#3e3f3a",
            "selected": "#49483e",
            "editor_bg": "#22231d",
            "editor_fg": "#f8f8f2",
        },
        "light": {
            "bg": "#f9f9f5",
            "sidebar": "#efefe9",
            "accent": "#7cb518",
            "accent_hover": "#0ea5e9",
            "text": "#272822",
            "muted": "#75715e",
            "card": "#ffffff",
            "selected": "#d9d9d3",
            "editor_bg": "#ffffff",
            "editor_fg": "#272822",
        },
    },
    "Nord": {
        "dark": {
            "bg": "#2e3440",
            "sidebar": "#3b4252",
            "accent": "#88c0d0",
            "accent_hover": "#81a1c1",
            "text": "#d8dee9",
            "muted": "#616e88",
            "card": "#434c5e",
            "selected": "#4c566a",
            "editor_bg": "#292e39",
            "editor_fg": "#d8dee9",
        },
        "light": {
            "bg": "#eceff4",
            "sidebar": "#e5e9f0",
            "accent": "#5e81ac",
            "accent_hover": "#4c6f99",
            "text": "#2e3440",
            "muted": "#616e88",
            "card": "#ffffff",
            "selected": "#d8dee9",
            "editor_bg": "#ffffff",
            "editor_fg": "#2e3440",
        },
    },
    "Gruvbox": {
        "dark": {
            "bg": "#282828",
            "sidebar": "#32302f",
            "accent": "#fabd2f",
            "accent_hover": "#fe8019",
            "text": "#ebdbb2",
            "muted": "#928374",
            "card": "#3c3836",
            "selected": "#504945",
            "editor_bg": "#242424",
            "editor_fg": "#ebdbb2",
        },
        "light": {
            "bg": "#fbf1c7",
            "sidebar": "#ebdbb2",
            "accent": "#b57614",
            "accent_hover": "#af3a03",
            "text": "#282828",
            "muted": "#928374",
            "card": "#ffffff",
            "selected": "#d5c4a1",
            "editor_bg": "#ffffff",
            "editor_fg": "#282828",
        },
    },
    "One Dark": {
        "dark": {
            "bg": "#282c34",
            "sidebar": "#31353f",
            "accent": "#61afef",
            "accent_hover": "#c678dd",
            "text": "#abb2bf",
            "muted": "#5c6370",
            "card": "#3a3f4b",
            "selected": "#4b5263",
            "editor_bg": "#23272e",
            "editor_fg": "#abb2bf",
        },
        "light": {
            "bg": "#fafafa",
            "sidebar": "#eaebed",
            "accent": "#4078f2",
            "accent_hover": "#a626a4",
            "text": "#383a42",
            "muted": "#6b7280",
            "card": "#ffffff",
            "selected": "#d1d5db",
            "editor_bg": "#ffffff",
            "editor_fg": "#383a42",
        },
    },
    "Material": {
        "dark": {
            "bg": "#263238",
            "sidebar": "#2f3d45",
            "accent": "#80cbc4",
            "accent_hover": "#ffcb6b",
            "text": "#eeffff",
            "muted": "#546e7a",
            "card": "#37474f",
            "selected": "#455a64",
            "editor_bg": "#212b30",
            "editor_fg": "#eeffff",
        },
        "light": {
            "bg": "#f5f5f5",
            "sidebar": "#e0e0e0",
            "accent": "#00897b",
            "accent_hover": "#f9a825",
            "text": "#263238",
            "muted": "#607d8b",
            "card": "#ffffff",
            "selected": "#bdbdbd",
            "editor_bg": "#ffffff",
            "editor_fg": "#263238",
        },
    },
    "Rose Pine": {
        "dark": {
            "bg": "#191724",
            "sidebar": "#1f1d2e",
            "accent": "#ebbcba",
            "accent_hover": "#f6c177",
            "text": "#e0def4",
            "muted": "#6e6a86",
            "card": "#26233a",
            "selected": "#2d2a45",
            "editor_bg": "#15131f",
            "editor_fg": "#e0def4",
        },
        "light": {
            "bg": "#faf4ed",
            "sidebar": "#f2e9e1",
            "accent": "#d7827e",
            "accent_hover": "#ea9d34",
            "text": "#575279",
            "muted": "#6e6a86",
            "card": "#ffffff",
            "selected": "#dfdad9",
            "editor_bg": "#ffffff",
            "editor_fg": "#575279",
        },
    },
    "Catppuccin": {
        "dark": {
            "bg": "#1e1e2e",
            "sidebar": "#252536",
            "accent": "#f5c2e7",
            "accent_hover": "#cba6f7",
            "text": "#cdd6f4",
            "muted": "#6c7086",
            "card": "#2a2a3e",
            "selected": "#35364a",
            "editor_bg": "#181825",
            "editor_fg": "#cdd6f4",
        },
        "light": {
            "bg": "#eff1f5",
            "sidebar": "#e6e9ef",
            "accent": "#ea76cb",
            "accent_hover": "#8839ef",
            "text": "#4c4f69",
            "muted": "#6c6f85",
            "card": "#ffffff",
            "selected": "#ccd0da",
            "editor_bg": "#ffffff",
            "editor_fg": "#4c4f69",
        },
    },
    "Retro Green": {
        "dark": {
            "bg": "#051a05",
            "sidebar": "#0a2e0a",
            "accent": "#00ff00",
            "accent_hover": "#39ff14",
            "text": "#00e600",
            "muted": "#006600",
            "card": "#0f3d0f",
            "selected": "#145214",
            "editor_bg": "#031403",
            "editor_fg": "#00ff00",
        },
        "light": {
            "bg": "#eaffea",
            "sidebar": "#c6ffc6",
            "accent": "#00cc00",
            "accent_hover": "#009900",
            "text": "#003300",
            "muted": "#339933",
            "card": "#ffffff",
            "selected": "#99ff99",
            "editor_bg": "#ffffff",
            "editor_fg": "#003300",
        },
    },
    "Retro Amber": {
        "dark": {
            "bg": "#1a0f00",
            "sidebar": "#2e1a00",
            "accent": "#ffb000",
            "accent_hover": "#ffcc00",
            "text": "#ffaa00",
            "muted": "#996600",
            "card": "#3d2400",
            "selected": "#523000",
            "editor_bg": "#140b00",
            "editor_fg": "#ffb000",
        },
        "light": {
            "bg": "#fff8e6",
            "sidebar": "#ffebb3",
            "accent": "#cc7a00",
            "accent_hover": "#995c00",
            "text": "#331e00",
            "muted": "#b37700",
            "card": "#ffffff",
            "selected": "#ffd966",
            "editor_bg": "#ffffff",
            "editor_fg": "#331e00",
        },
    },
    "Retro Cyan": {
        "dark": {
            "bg": "#001a1a",
            "sidebar": "#002e2e",
            "accent": "#00ffff",
            "accent_hover": "#00e5e5",
            "text": "#00cccc",
            "muted": "#006666",
            "card": "#003d3d",
            "selected": "#005252",
            "editor_bg": "#001414",
            "editor_fg": "#00ffff",
        },
        "light": {
            "bg": "#e6ffff",
            "sidebar": "#b3ffff",
            "accent": "#009999",
            "accent_hover": "#006666",
            "text": "#003333",
            "muted": "#008080",
            "card": "#ffffff",
            "selected": "#66ffff",
            "editor_bg": "#ffffff",
            "editor_fg": "#003333",
        },
    },
    "Matrix": {
        "dark": {
            "bg": "#000000",
            "sidebar": "#001100",
            "accent": "#00ff41",
            "accent_hover": "#00cc33",
            "text": "#00ff41",
            "muted": "#008f11",
            "card": "#002200",
            "selected": "#003300",
            "editor_bg": "#000000",
            "editor_fg": "#00ff41",
        },
        "light": {
            "bg": "#e6ffe6",
            "sidebar": "#ccffcc",
            "accent": "#00b300",
            "accent_hover": "#008000",
            "text": "#003300",
            "muted": "#009900",
            "card": "#ffffff",
            "selected": "#80ff80",
            "editor_bg": "#ffffff",
            "editor_fg": "#003300",
        },
    },
    "Blood Red": {
        "dark": {
            "bg": "#1a0000",
            "sidebar": "#2e0000",
            "accent": "#ff0000",
            "accent_hover": "#ff3333",
            "text": "#ff6666",
            "muted": "#990000",
            "card": "#3d0000",
            "selected": "#520000",
            "editor_bg": "#140000",
            "editor_fg": "#ff0000",
        },
        "light": {
            "bg": "#ffe6e6",
            "sidebar": "#ffcccc",
            "accent": "#cc0000",
            "accent_hover": "#990000",
            "text": "#330000",
            "muted": "#b30000",
            "card": "#ffffff",
            "selected": "#ff9999",
            "editor_bg": "#ffffff",
            "editor_fg": "#330000",
        },
    },
    "Banana": {
        "dark": {
            "bg": "#1a1a00",
            "sidebar": "#2e2e00",
            "accent": "#ffff00",
            "accent_hover": "#ffff33",
            "text": "#ffff66",
            "muted": "#999900",
            "card": "#3d3d00",
            "selected": "#525200",
            "editor_bg": "#141400",
            "editor_fg": "#ffff00",
        },
        "light": {
            "bg": "#ffffe6",
            "sidebar": "#ffffb3",
            "accent": "#b3b300",
            "accent_hover": "#808000",
            "text": "#333300",
            "muted": "#999900",
            "card": "#ffffff",
            "selected": "#ffff66",
            "editor_bg": "#ffffff",
            "editor_fg": "#333300",
        },
    },
    "Bubblegum": {
        "dark": {
            "bg": "#2a0a1a",
            "sidebar": "#3d0f26",
            "accent": "#ff69b4",
            "accent_hover": "#ff85c2",
            "text": "#ffcce6",
            "muted": "#993366",
            "card": "#4a1430",
            "selected": "#5c1a3d",
            "editor_bg": "#220815",
            "editor_fg": "#ff69b4",
        },
        "light": {
            "bg": "#fff0f7",
            "sidebar": "#ffd6e8",
            "accent": "#d63384",
            "accent_hover": "#a61e66",
            "text": "#33001f",
            "muted": "#b3477a",
            "card": "#ffffff",
            "selected": "#ffb3d9",
            "editor_bg": "#ffffff",
            "editor_fg": "#33001f",
        },
    },
    "Grape": {
        "dark": {
            "bg": "#1a0a2e",
            "sidebar": "#260f45",
            "accent": "#bf00ff",
            "accent_hover": "#d24dff",
            "text": "#e6b3ff",
            "muted": "#660099",
            "card": "#33154d",
            "selected": "#421a63",
            "editor_bg": "#140620",
            "editor_fg": "#bf00ff",
        },
        "light": {
            "bg": "#f5e6ff",
            "sidebar": "#e6ccff",
            "accent": "#7a00cc",
            "accent_hover": "#5c0099",
            "text": "#26004d",
            "muted": "#8a4dbf",
            "card": "#ffffff",
            "selected": "#cc99ff",
            "editor_bg": "#ffffff",
            "editor_fg": "#26004d",
        },
    },
}


    def __init__(self, root):
        self.root = root
        self.root.title("Fusion - Manage your server posts")
        self.root.geometry("1100x700")
        self.root.minsize(800, 500)

        self.theme_name, self.mode, self.show_intro = self._load_settings()
        self.colors = self.THEMES[self.theme_name][self.mode].copy()

        self.announcements = [
            {"name": "Welcome to Fusion", "content": "Welcome to fusion!\n\nInstructions on how to use Fusion can be found [here](https://github.com/zion8992/ironite/blob/main/fusion/Fusion_Setup.m)\n\nYou can change the size of the preview to make it larger if needed.",
             "date": go_format_now()},  # >>> added date field
        ]

        self.current_index = None
        self.editor_open = False

        self._setup_styles()
        self._build_ui()

        self.root.protocol("WM_DELETE_WINDOW", self._on_close)

        if self.show_intro:
            self._show_intro()
        else:
            self.main_frame.pack(fill="both", expand=True)

    def _load_settings(self):
        config = configparser.ConfigParser()
        if os.path.exists(CONFIG_FILE):
            config.read(CONFIG_FILE)
        else:
            config["Settings"] = {}

        theme = config["Settings"].get("theme", "Midnight")
        mode = config["Settings"].get("mode", "dark").lower()
        show_intro = config["Settings"].getboolean("show_intro", True)

        if theme not in self.THEMES or mode not in self.THEMES[theme]:
            theme = "Midnight"
            mode = "dark"

        return theme, mode, show_intro

    def _save_settings(self):
        config = configparser.ConfigParser()
        config["Settings"] = {
            "theme": self.theme_name,
            "mode": self.mode,
            "show_intro": str(self.show_intro),
        }
        with open(CONFIG_FILE, "w", encoding="utf-8") as f:
            config.write(f)

    def _on_close(self):
        self._save_settings()
        self.root.destroy()

    def _setup_styles(self):
        style = ttk.Style()
        style.theme_use("clam")

        style.configure("Sidebar.TFrame", background=self.colors["sidebar"])
        style.configure("Main.TFrame", background=self.colors["bg"])

        style.configure(
            "Accent.TButton",
            background=self.colors["accent"],
            foreground="#1e1e2e",
            font=("Inter", 11, "bold"),
            borderwidth=0,
            focusthickness=0,
            padding=10,
        )
        style.map("Accent.TButton", background=[("active", self.colors["accent_hover"])])

        style.configure(
            "Toolbar.TButton",
            background=self.colors["card"],
            foreground=self.colors["text"],
            font=("Inter", 10),
            borderwidth=0,
            focusthickness=0,
            padding=6,
        )
        style.map("Toolbar.TButton", background=[("active", self.colors["selected"])])

        style.configure(
            "SidebarAction.TButton",
            background=self.colors["card"],
            foreground=self.colors["text"],
            font=("Inter", 10),
            borderwidth=0,
            focusthickness=0,
            padding=8,
        )
        style.map("SidebarAction.TButton", background=[("active", self.colors["selected"])])

        style.configure("TSeparator", background=self.colors["selected"])

    def _build_ui(self):
        self.main_frame = tk.Frame(self.root, bg=self.colors["bg"])

        main_pane = tk.PanedWindow(self.main_frame, orient="horizontal", sashwidth=4, bg=self.colors["bg"])
        main_pane.pack(fill="both", expand=True)

        # Sidebar
        self.sidebar = tk.Frame(main_pane, bg=self.colors["sidebar"], width=260)
        self.sidebar.pack_propagate(False)
        main_pane.add(self.sidebar, minsize=220)

        create_btn = ttk.Button(
            self.sidebar,
            text="+ Create Announcement",
            style="Accent.TButton",
            command=self._create_announcement,
        )
        create_btn.pack(fill="x", padx=16, pady=16)

        self.announcements_label = tk.Label(
            self.sidebar,
            text="ANNOUNCEMENTS",
            bg=self.colors["sidebar"],
            fg=self.colors["muted"],
            font=("Inter", 9, "bold"),
        )
        self.announcements_label.pack(anchor="w", padx=16, pady=(0, 8))

        self.listbox = tk.Listbox(
            self.sidebar,
            bg=self.colors["sidebar"],
            fg=self.colors["text"],
            selectbackground=self.colors["selected"],
            selectforeground=self.colors["text"],
            font=("Inter", 11),
            borderwidth=0,
            highlightthickness=0,
            activestyle="none",
        )
        self.listbox.pack(fill="both", expand=True, padx=16, pady=(0, 8))
        self.listbox.bind("<<ListboxSelect>>", self._on_select)
        self.listbox.bind("<Button-3>", self._show_context_menu)
        self.listbox.bind("<Double-Button-1>", self._on_double_click)
        self.listbox.bind("<Delete>", lambda e: self._delete_announcement())

        # Sidebar bottom actions
        ttk.Separator(self.sidebar, orient="horizontal").pack(fill="x", padx=16, pady=8)

        import_btn = ttk.Button(
            self.sidebar,
            text="Import Feed",
            style="SidebarAction.TButton",
            command=self._import_feed,
        )
        import_btn.pack(fill="x", padx=16, pady=(0, 8))

        export_btn = ttk.Button(
            self.sidebar,
            text="Export Feed",
            style="SidebarAction.TButton",
            command=self._export_feed,
        )
        export_btn.pack(fill="x", padx=16, pady=(0, 8))

        settings_btn = ttk.Button(
            self.sidebar,
            text="Settings",
            style="Accent.TButton",
            command=self._open_settings,
        )
        settings_btn.pack(fill="x", padx=16, pady=(0, 16))

        # Center
        self.center = tk.Frame(main_pane, bg=self.colors["bg"])
        main_pane.add(self.center, minsize=400)

        self._show_placeholder()
        self._refresh_list()

    def _show_intro(self):
        self.intro = tk.Frame(self.root, bg=self.colors["bg"])
        self.intro.pack(fill="both", expand=True)

        self.intro_label = tk.Label(
            self.intro,
            text="Fusion",
            bg=self.colors["bg"],
            fg=self.colors["bg"],
            font=("Inter", 70, "bold"),
        )
        self.intro_label.place(relx=0.5, rely=0.45, anchor="center")

        self._animate_intro(0)

    def _animate_intro(self, step):
        if step > INTRO_FADE_STEPS:
            self.intro.after(INTRO_DURATION_MS, self._finish_intro)
            return

        ratio = step / INTRO_FADE_STEPS
        color = self._blend_color(self.colors["bg"], self.colors["accent"], ratio)
        self.intro_label.configure(fg=color)

        delay = INTRO_DURATION_MS // (INTRO_FADE_STEPS * 2)
        self.intro.after(delay, lambda: self._animate_intro(step + 1))

    def _finish_intro(self):
        self.intro.destroy()
        self.main_frame.pack(fill="both", expand=True)

    def _blend_color(self, c1, c2, ratio):
        r1, g1, b1 = int(c1[1:3], 16), int(c1[3:5], 16), int(c1[5:7], 16)
        r2, g2, b2 = int(c2[1:3], 16), int(c2[3:5], 16), int(c2[5:7], 16)
        r = int(r1 + (r2 - r1) * ratio)
        g = int(g1 + (g2 - g1) * ratio)
        b = int(b1 + (b2 - b1) * ratio)
        return f"#{r:02x}{g:02x}{b:02x}"

    def _refresh_list(self):
        self.listbox.delete(0, "end")
        for item in self.announcements:
            self.listbox.insert("end", item["name"])

    def _selected_index(self):
        selection = self.listbox.curselection()
        return selection[0] if selection else None

    def _show_placeholder(self):
        for child in self.center.winfo_children():
            child.destroy()

        self.editor_open = False
        self.current_index = None

        self.placeholder = tk.Label(
            self.center,
            text="Double-click an announcement to edit it",
            bg=self.colors["bg"],
            fg=self.colors["muted"],
            font=("Inter", 14),
        )
        self.placeholder.place(relx=0.5, rely=0.5, anchor="center")

    def _set_editor_split(self, pane):
        pane.update_idletasks()
        width = pane.winfo_width()
        pane.sash_place(0, int(width * 0.55), 0)

    def _show_editor(self, index=None):
        for child in self.center.winfo_children():
            child.destroy()

        self.editor_open = True
        self.current_index = index

        # Toolbar
        self.toolbar = tk.Frame(self.center, bg=self.colors["card"], height=44)
        self.toolbar.pack(fill="x", side="top")
        self.toolbar.pack_propagate(False)

        title = self.announcements[index]["name"] if index is not None else ""
        self.title_label = tk.Label(
            self.toolbar,
            text="Title",
            bg=self.colors["card"],
            fg=self.colors["muted"],
            font=("Inter", 10),
        )
        self.title_label.pack(side="left", padx=(12, 6))

        self.title_entry = tk.Entry(
            self.toolbar,
            bg=self.colors["editor_bg"],
            fg=self.colors["text"],
            insertbackground=self.colors["text"],
            font=("Inter", 11),
            relief="flat",
            highlightthickness=1,
            highlightcolor=self.colors["accent"],
            highlightbackground=self.colors["selected"],
        )
        self.title_entry.pack(side="left", fill="y", pady=6)
        self.title_entry.insert(0, title)

        ttk.Separator(self.toolbar, orient="vertical").pack(side="left", fill="y", padx=10, pady=8)

        buttons = [
            ("Bold", self._make_bold),
            ("Italic", self._make_italic),
            ("Heading", self._make_heading),
            ("Link", self._make_link),
            ("List", self._make_list),
        ]
        for label, command in buttons:
            btn = ttk.Button(self.toolbar, text=label, style="Toolbar.TButton", command=command)
            btn.pack(side="left", padx=2)

        ttk.Separator(self.toolbar, orient="vertical").pack(side="left", fill="y", padx=10, pady=8)

        save_btn = ttk.Button(self.toolbar, text="Save", style="Toolbar.TButton", command=self._save)
        save_btn.pack(side="left", padx=2)

        close_btn = ttk.Button(self.toolbar, text="Close", style="Toolbar.TButton", command=self._show_placeholder)
        close_btn.pack(side="right", padx=12)

        # Editor + preview pane
        content_pane = tk.PanedWindow(
            self.center,
            orient="horizontal",
            sashwidth=6,
            sashrelief="raised",
            showhandle=True,
            bg=self.colors["selected"],
            bd=0,
        )
        content_pane.pack(fill="both", expand=True)

        editor_frame = tk.Frame(content_pane, bg=self.colors["bg"])
        content_pane.add(editor_frame, minsize=200)

        self.editor_label = tk.Label(
            editor_frame,
            text="Markdown",
            bg=self.colors["bg"],
            fg=self.colors["muted"],
            font=("Inter", 9, "bold"),
        )
        self.editor_label.pack(anchor="w", padx=12, pady=(8, 4))

        self.editor_text = tk.Text(
            editor_frame,
            bg=self.colors["editor_bg"],
            fg=self.colors["editor_fg"],
            insertbackground=self.colors["text"],
            font=("JetBrains Mono", 12),
            wrap="word",
            relief="flat",
            borderwidth=0,
            padx=10,
            pady=10,
            undo=True,
        )
        self.editor_text.pack(fill="both", expand=True, padx=12, pady=(0, 12))
        self.editor_text.bind("<KeyRelease>", lambda e: self._update_preview())

        content = self.announcements[index]["content"] if index is not None else ""
        self.editor_text.insert("1.0", content)

        preview_frame = tk.Frame(content_pane, bg=self.colors["bg"])
        content_pane.add(preview_frame, minsize=200)

        self.preview_label = tk.Label(
            preview_frame,
            text="Preview",
            bg=self.colors["bg"],
            fg=self.colors["muted"],
            font=("Inter", 9, "bold"),
        )
        self.preview_label.pack(anchor="w", padx=12, pady=(8, 4))

        if HtmlFrame is not None:
            self.preview = HtmlFrame(preview_frame, messages_enabled=False)
            self.preview.pack(fill="both", expand=True, padx=12, pady=(0, 12))
        else:
            self.preview = tk.Text(
                preview_frame,
                bg=self.colors["editor_bg"],
                fg=self.colors["editor_fg"],
                font=("JetBrains Mono", 11),
                wrap="word",
                relief="flat",
                borderwidth=0,
                padx=10,
                pady=10,
                state="disabled",
            )
            self.preview.pack(fill="both", expand=True, padx=12, pady=(0, 12))

        content_pane.after_idle(lambda: self._set_editor_split(content_pane))

        self._update_preview()

    def _update_preview(self):
        md = self.editor_text.get("1.0", "end").strip()
        html = markdown.markdown(md, extensions=["fenced_code", "tables"])
        styled = f"""<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body {{ background: {self.colors['bg']}; color: {self.colors['text']}; font-family: sans-serif; padding: 1rem; line-height: 1.5; }}
  h1, h2, h3, h4, h5, h6 {{ color: {self.colors['accent']}; }}
  a {{ color: {self.colors['accent_hover']}; }}
  code {{ background: {self.colors['card']}; padding: 2px 4px; border-radius: 4px; }}
  pre {{ background: {self.colors['card']}; padding: 10px; border-radius: 6px; overflow-x: auto; }}
  blockquote {{ border-left: 3px solid {self.colors['accent']}; margin-left: 0; padding-left: 1rem; color: {self.colors['muted']}; }}
  table {{ border-collapse: collapse; width: 100%; }}
  th, td {{ border: 1px solid {self.colors['selected']}; padding: 6px; }}
  th {{ background: {self.colors['card']}; }}
</style>
</head>
<body>
{html}
</body>
</html>"""
        if HtmlFrame is not None:
            self.preview.load_html(styled)
        else:
            self.preview.config(state="normal")
            self.preview.delete("1.0", "end")
            self.preview.insert("1.0", styled)
            self.preview.config(state="disabled")

    # Toolbar actions
    def _wrap_selection(self, before, after):
        try:
            start = self.editor_text.index("sel.first")
            end = self.editor_text.index("sel.last")
            selected = self.editor_text.get(start, end)
            self.editor_text.delete(start, end)
            self.editor_text.insert(start, f"{before}{selected}{after}")
            self.editor_text.mark_set("insert", f"{start}+{len(before)}c")
        except tk.TclError:
            self.editor_text.insert("insert", f"{before}{after}")
            self.editor_text.mark_set("insert", "insert-{}c".format(len(after)))
        self.editor_text.focus_set()
        self._update_preview()

    def _make_bold(self):
        self._wrap_selection("**", "**")

    def _make_italic(self):
        self._wrap_selection("*", "*")

    def _make_heading(self):
        line = self.editor_text.index("insert").split(".")[0]
        current = self.editor_text.get(f"{line}.0", f"{line}.end")
        if not current.startswith("## "):
            self.editor_text.insert(f"{line}.0", "## ")
        self._update_preview()

    def _make_link(self):
        url = simpledialog.askstring("Link", "Enter URL:", parent=self.root)
        if not url:
            return
        try:
            start = self.editor_text.index("sel.first")
            end = self.editor_text.index("sel.last")
            text = self.editor_text.get(start, end) or "link"
            self.editor_text.delete(start, end)
            self.editor_text.insert(start, f"[{text}]({url})")
        except tk.TclError:
            self.editor_text.insert("insert", f"[link]({url})")
        self.editor_text.focus_set()
        self._update_preview()

    def _make_list(self):
        line = self.editor_text.index("insert").split(".")[0]
        current = self.editor_text.get(f"{line}.0", f"{line}.end")
        if not current.startswith("- "):
            self.editor_text.insert(f"{line}.0", "- ")
        self._update_preview()

    # CRUD
    def _create_announcement(self):
        self._show_editor(index=None)

    def _on_double_click(self, event):
        idx = self._selected_index()
        if idx is not None:
            self._show_editor(index=idx)

    def _on_select(self, event=None):
        pass

    def _save(self):
        name = self.title_entry.get().strip()
        content = self.editor_text.get("1.0", "end").strip()

        if not name:
            messagebox.showwarning("Invalid", "Title cannot be empty.")
            return

        if self.current_index is None:
            self.announcements.append({
                "name": name,
                "content": content,
                "date": go_format_now(),  # >>> added date field
            })
        else:
            # Preserve the existing date on edits; only new announcements
            # get a freshly stamped date at creation time.
            existing_date = self.announcements[self.current_index].get("date", go_format_now())
            self.announcements[self.current_index] = {
                "name": name,
                "content": content,
                "date": existing_date,  # >>> added date field
            }

        self._refresh_list()

        target = len(self.announcements) - 1 if self.current_index is None else self.current_index
        self.listbox.selection_clear(0, "end")
        self.listbox.selection_set(target)
        self.listbox.see(target)
        self.current_index = target

    def _edit_announcement(self):
        idx = self._selected_index()
        if idx is not None:
            self._show_editor(index=idx)

    def _delete_announcement(self):
        idx = self._selected_index()
        if idx is None:
            return

        name = self.announcements[idx]["name"]
        if messagebox.askyesno("Delete", f"Delete '{name}'?"):
            del self.announcements[idx]
            self._refresh_list()

            if self.editor_open and self.current_index == idx:
                self._show_placeholder()
            elif self.editor_open and self.current_index is not None and self.current_index > idx:
                self.current_index -= 1

    def _show_context_menu(self, event):
        self.listbox.selection_clear(0, "end")
        index = self.listbox.nearest(event.y)
        if index >= 0:
            self.listbox.selection_set(index)
            menu = tk.Menu(self.root, tearoff=0, bg=self.colors["card"], fg=self.colors["text"])
            menu.add_command(label="Open", command=self._edit_announcement)
            menu.add_command(label="Edit", command=self._edit_announcement)
            menu.add_command(label="Delete", command=self._delete_announcement)
            menu.post(event.x_root, event.y_root)

    # Import / Export
    def _export_feed(self):
        if not self.announcements:
            messagebox.showinfo("Export", "No announcements to export.")
            return

        path = filedialog.asksaveasfilename(
            defaultextension=".xml",
            filetypes=[("XML Feed", "*.xml"), ("All Files", "*.*")],
            title="Export Announcement Feed",
        )
        if not path:
            return

        try:
            root = ET.Element("feed")
            for item in self.announcements:
                ann = ET.SubElement(root, "announcement")
                title = ET.SubElement(ann, "title")
                title.text = item["name"]
                content = ET.SubElement(ann, "content")
                content.text = item["content"]
                # >>> write the date field to the exported XML
                date_elem = ET.SubElement(ann, "date")
                date_elem.text = item.get("date", "")

            tree = ET.ElementTree(root)
            ET.indent(tree, space="  ")
            tree.write(path, encoding="utf-8", xml_declaration=True)

            messagebox.showinfo("Export Successful", f"Saved {len(self.announcements)} announcements to:\n{path}")
        except Exception as e:
            messagebox.showerror("Export Failed", str(e))

    def _import_feed(self):
        path = filedialog.askopenfilename(
            defaultextension=".xml",
            filetypes=[("XML Feed", "*.xml"), ("All Files", "*.*")],
            title="Import Announcement Feed",
        )
        if not path:
            return

        try:
            tree = ET.parse(path)
            root = tree.getroot()

            imported = []
            for ann in root.findall("announcement"):
                title_elem = ann.find("title")
                content_elem = ann.find("content")
                date_elem = ann.find("date")  # >>> read the date element if present
                title = (title_elem.text or "").strip() if title_elem is not None else ""
                content = (content_elem.text or "") if content_elem is not None else ""
                date = (date_elem.text or "").strip() if date_elem is not None and date_elem.text else ""

                if title:
                    imported.append({
                        "name": title,
                        "content": content,
                        "date": date,  # >>> added date field
                    })

            if not imported:
                messagebox.showwarning("Import", "Invalid announcement file.")
                return

            if self.announcements:
                response = messagebox.askyesno(
                    "Import",
                    f"Found {len(imported)} announcements. This will override the current list."
                )
                if not response:
                    return

            self.announcements = imported
            self._refresh_list()
            self._show_placeholder()

            messagebox.showinfo("Import Successful", f"Imported {len(imported)} announcements.")
        except ET.ParseError as e:
            messagebox.showerror("Import Failed", f"Invalid XML:\n{e}")
        except Exception as e:
            messagebox.showerror("Import Failed", str(e))

    # Settings
    def _open_settings(self):
        window = tk.Toplevel(self.root)
        window.title("Settings")
        window.geometry("320x300")
        window.configure(bg=self.colors["bg"])
        window.transient(self.root)
        window.grab_set()

        tk.Label(
            window,
            text="Theme",
            bg=self.colors["bg"],
            fg=self.colors["text"],
            font=("Inter", 11),
        ).pack(anchor="w", padx=16, pady=(16, 4))

        theme_var = tk.StringVar(value=self.theme_name)
        theme_combo = ttk.Combobox(
            window,
            values=list(self.THEMES.keys()),
            textvariable=theme_var,
            state="readonly",
            font=("Inter", 10),
        )
        theme_combo.pack(fill="x", padx=16, pady=(0, 12))

        tk.Label(
            window,
            text="Mode",
            bg=self.colors["bg"],
            fg=self.colors["text"],
            font=("Inter", 11),
        ).pack(anchor="w", padx=16, pady=(0, 4))

        mode_var = tk.StringVar(value=self.mode.capitalize())
        mode_combo = ttk.Combobox(
            window,
            values=["Dark", "Light"],
            textvariable=mode_var,
            state="readonly",
            font=("Inter", 10),
        )
        mode_combo.pack(fill="x", padx=16, pady=(0, 12))

        intro_var = tk.BooleanVar(value=self.show_intro)
        intro_check = tk.Checkbutton(
            window,
            text="Show intro animation on startup",
            variable=intro_var,
            bg=self.colors["bg"],
            fg=self.colors["text"],
            selectcolor=self.colors["card"],
            activebackground=self.colors["bg"],
            activeforeground=self.colors["text"],
            font=("Inter", 10),
            anchor="w",
        )
        intro_check.pack(fill="x", padx=16, pady=(0, 16))

        def apply_settings():
            self._set_theme(theme_var.get(), mode_var.get().lower())
            self.show_intro = intro_var.get()
            window.destroy()

        apply_btn = ttk.Button(
            window,
            text="Apply",
            style="Accent.TButton",
            command=apply_settings,
        )
        apply_btn.pack(fill="x", padx=16, pady=(0, 8))

        cancel_btn = ttk.Button(
            window,
            text="Cancel",
            style="SidebarAction.TButton",
            command=window.destroy,
        )
        cancel_btn.pack(fill="x", padx=16, pady=(0, 16))

    def _set_theme(self, theme_name, mode):
        if theme_name not in self.THEMES or mode not in self.THEMES[theme_name]:
            return

        self.theme_name = theme_name
        self.mode = mode
        self.colors = self.THEMES[theme_name][mode].copy()
        self._apply_theme()

    def _apply_theme(self):
        self._setup_styles()

        self.root.configure(bg=self.colors["bg"])
        self.sidebar.configure(bg=self.colors["sidebar"])
        self.announcements_label.configure(bg=self.colors["sidebar"], fg=self.colors["muted"])

        self.listbox.configure(
            bg=self.colors["sidebar"],
            fg=self.colors["text"],
            selectbackground=self.colors["selected"],
            selectforeground=self.colors["text"],
        )

        self.center.configure(bg=self.colors["bg"])

        if not self.editor_open:
            self.placeholder.configure(bg=self.colors["bg"], fg=self.colors["muted"])
        else:
            self.toolbar.configure(bg=self.colors["card"])
            self.title_label.configure(bg=self.colors["card"], fg=self.colors["muted"])
            self.title_entry.configure(
                bg=self.colors["editor_bg"],
                fg=self.colors["text"],
                insertbackground=self.colors["text"],
                highlightcolor=self.colors["accent"],
                highlightbackground=self.colors["selected"],
            )
            self.editor_label.configure(bg=self.colors["bg"], fg=self.colors["muted"])
            self.preview_label.configure(bg=self.colors["bg"], fg=self.colors["muted"])
            self.editor_text.configure(
                bg=self.colors["editor_bg"],
                fg=self.colors["editor_fg"],
                insertbackground=self.colors["text"],
            )

            if isinstance(self.preview, tk.Text):
                self.preview.configure(
                    bg=self.colors["editor_bg"],
                    fg=self.colors["editor_fg"],
                )

            self._update_preview()


def main():
    root = tk.Tk()
    app = AnnouncementApp(root)
    root.mainloop()


if __name__ == "__main__":
    main()
