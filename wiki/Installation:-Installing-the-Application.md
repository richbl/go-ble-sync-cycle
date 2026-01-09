<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

Once the **BLE Sync Cycle** application binary has been built, it's ready to be installed.

For running **BLE Sync Cycle** in CLI mode, there is no additional work to be done: you can run the application locally from the command line with various configuration flags (see [Basic Usage: Using the Command Line Options](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-CLI-Mode)).

For running **BLE Sync Cycle** in GUI mode, the installation process requires the following files to be installed:

- The `ble-sync-cycle` executable (moved to a system-default binary file location)
- The **BLE Sync Cycle** icon file (`com.github.richbl.ble-sync-cycle.svg`)  
- The `com.github.richbl.ble-sync-cycle.desktop` file

> The `.desktop` file is a [standardized text file in Linux/Unix systems](https://specifications.freedesktop.org/desktop-entry/latest/index.html) that acts as a launcher for applications, defining their name, icon, menu location, and the command(s) to run them, enabling integration with the operating system's desktop environment (DE).

In addition, the creation of a local folder is needed for storing/managing BSC TOML configuration files (the session files representing cycling events).

### GUI Mode Installation

1. Copy the `ble-sync-cycle` executable to a directory in your path (e.g., `~/.local/bin`) Note that on some systems, this directory may not exist, in which case you may first need to create it:

    ```console
    mkdir -p ~/.local/bin
    ```

    Then copy the executable to that directory:

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
    cp ui/assets/com.github.richbl.ble-sync-cycle.svg ~/.local/share/icons/hicolor/scalable/apps
    ```

1. Copy the `com.github.richbl.ble-sync-cycle.desktop` project file (found in the `/ui/assets` directory) to the local applications directory (e.g., `~/.local/share/applications`):

    ```console
    cp ui/assets/com.github.richbl.ble-sync-cycle.desktop ~/.local/share/applications
    ```

1. Create a new folder called `com.github.richbl.ble-sync-cycle` in`~/.config`:

    ```console
    mkdir -p ~/.config/com.github.richbl.ble-sync-cycle
    ```

    > This folder is important, as its the location where **BLE Sync Cycle** looks for application configuration files (called BSC TOML files).
