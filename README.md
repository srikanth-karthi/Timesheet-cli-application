
# 🕒 Timesheet CLI Application

Track, manage, and report your timesheets directly from the terminal — no local storage, no hassle.  
Everything is backed by **Google Sheets** so your team stays in sync and transparent.

---

## 📦 Features

- ⏱️ Real-time start/stop work sessions
- 🧾 Manually log tasks with hours and descriptions
- 📅 Weekly reports grouped by project
- 🧠 Bucket/project switching
- ☁️ All logs stored in a shared **Google Sheet**

---

## 🚀 Quick Start

### 📥 1. Download the Latest Release

| OS        | Binary Link |
|-----------|-------------|
| 🐧 Linux   | [Download `timesheet-linux`](https://github.com/srikanth-karthi/Timesheet-cli-application/releases/latest/download/timesheet-linux) |
| 🍎 macOS   | [Download `timesheet-mac`](https://github.com/srikanth-karthi/Timesheet-cli-application/releases/latest/download/timesheet-mac-arm64) |
| 🪟 Windows | [Download `timesheet.exe`](https://github.com/srikanth-karthi/Timesheet-cli-application/releases/latest/download/timesheet.exe) |

> 🔄 Always grab the latest version from the [Releases Page](https://github.com/srikanth-karthi/Timesheet-cli-application/releases).

---

### 🛠️ 2. Install the CLI

#### 🐧 Linux / 🍎 macOS:

```bash
curl -L https://github.com/srikanth-karthi/Timesheet-cli-application/releases/latest/download/timesheet-mac-arm64 -o timesheet
chmod +x timesheet
sudo mv timesheet /usr/local/bin/timesheet

sudo mkdir -p /usr/local/share/zsh/site-functions

# Install the completion script here
/usr/local/bin/timesheet completion zsh | sudo tee /usr/local/share/zsh/site-functions/_timesheet > /dev/null
```

#### 🪟 Windows:

1. Download `timesheet.exe`  
2. *(Optional)* Add the folder containing the binary to your **System PATH**
3. Run the CLI using:

```powershell
.\timesheet.exe
```

---

## 🧑‍💻 Command Usage

```
timesheet

Track, manage, and report timesheets directly from the terminal.

Usage:
  timesheet [command]

Available Commands:
  bucket      List or switch buckets
  help        Help about any command
  list        List all buckets (shows current)
  log         📝 Manually log a task with hours
  new         Create or switch to a bucket
  report      📊 Show this week's summary grouped by project
  setup       Authenticate and set up your timesheet
  start       ⏱️ Start tracking time
  stop        ⏹️ Stop tracking the current session and log the duration

Flags:
  -h, --help     Help for timesheet
  -t, --toggle   Help message for toggle
```

---

📣 **Note**: First-time users must run `timesheet setup` to authenticate and link their Google Sheet.

---

## 📌 License

MIT © [Srikanth K](https://github.com/srikanth-karthi)


