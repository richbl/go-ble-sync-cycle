<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

Once the **BLE Sync Cycle** application has been built, it's ready to be installed etiher as a user-level application (recommended) or as a system-level application.

There are three files that will need to be installed:

- The `ble-sync-cycle` executable  
- The **BLE Sync Cycle** icon file (`com.github.richbl.ble-sync-cycle.svg`)  
- The `com.github.richbl.ble-sync-cycle.desktop` file

> The `.desktop` file is a [standardized text file in Linux/Unix systems](https://specifications.freedesktop.org/desktop-entry/latest/index.html) that acts as a launcher for applications, defining their name, icon, menu location, and the command(s) to run them, enabling integration with the operating system's desktop environment (DE).

### User-Level Installation (Recommended)

1. Copy the `ble-sync-cycle` executable to a directory in your path (e.g., `~/.local/bin`):

    ```console
    cp ble-sync-cycle ~/.local/bin
    ```

1. Copy the **BLE Sync Cycle** icon file (`com.github.richbl.ble-sync-cycle.svg` found in the`/ui/assets` directory) to the local icons directory (e.g., `~/.local/share/icons/hicolor/scalable/apps
`). Note on some systems, this directory may not exist, in which case you may first need to create it:

    ```console
    mkdir -p ~/.local/share/icons/hicolor/scalable/apps
    ```

    Then copy the icon to that directory:

    ```console
    cp com.github.richbl.ble-sync-cycle.svg ~/.local/share/icons/hicolor/scalable/apps
    ```

1. Copy the `com.github.richbl.ble-sync-cycle.desktop` project file (found in the `/ui/assets` directory) to the local applications directory (e.g., `~/.local/share/applications`):

    ```console
    cp com.github.richbl.ble-sync-cycle.desktop ~/.local/share/applications
    ```

### System-Level Installation

> **Note:** This installation method is not recommended for most users, since it requires root privileges to install the application and supporting files.

1. Copy the `ble-sync-cycle` executable to the application binary directory (e.g., `/usr/local/bin`):

    ```console
    sudo cp ble-sync-cycle /usr/local/bin
    ```

1. Copy the **BLE Sync Cycle** icon file (`com.github.richbl.ble-sync-cycle.svg` found in the`/ui/assets` directory) to the system icons directory (e.g., `/usr/share/icons/hicolor/scalable/apps`).

    ```console
    sudo cp com.github.richbl.ble-sync-cycle.svg /usr/share/icons/hicolor/scalable/apps
    ```

1. Copy the `com.github.richbl.ble-sync-cycle.desktop` project file (found in the `/ui/assets` directory) to the system applications directory (e.g., `/usr/share/applications`):

    ```console
    sudo cp com.github.richbl.ble-sync-cycle.desktop /usr/share/applications
    ```
