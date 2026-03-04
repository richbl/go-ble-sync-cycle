<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

Once the **BLE Sync Cycle** application binary has been downloaded or built locally, it's ready to be installed.

### CLI Mode Installation

To run **BLE Sync Cycle** in CLI mode, there is no additional work to be done: you can run the application locally from the command line with various configuration flags (see [Basic Usage: Using the Command Line Options](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-CLI-Mode)).

### GUI Mode Installation

When running **BLE Sync Cycle** in GUI mode, the installation process requires that the **BLE Sync Cycle** application run it's internal installer/uninstaller.

To install **BLE Sync Cycle**, run the application binary with the `--install` flag. For example:

  ```bash
  ./ble-sync-cycle --install
  ```

A successful installation will result in the following output:

```console
18:23:43 [INF] [APP] ---------------------------------------------------
18:23:43 [INF] [APP] BLE Sync Cycle v0.60.0 starting...
18:23:43 [INF] [APP] ---------------------------------------------------

Installing the following BLE Sync Cycle v0.60.0 files...

Binary:       /home/richbl/.local/bin/ble-sync-cycle
Desktop file: /home/richbl/.local/share/applications/com.github.richbl.ble-sync-cycle.desktop
Icon:         /home/richbl/.local/share/icons/hicolor/scalable/apps/com.github.richbl.ble-sync-cycle.svg

Installation completed successfully.

18:23:43 [INF] [APP] ---------------------------------------------------
18:23:43 [INF] [APP] BLE Sync Cycle v0.60.0 shutdown complete. Goodbye
18:23:43 [INF] [APP] ---------------------------------------------------
```

### GUI Mode Uninstallation

To uninstall **BLE Sync Cycle**, run the application binary with the `--uninstall` flag. For example:

```bash
./ble-sync-cycle --uninstall
```

A successful uninstallation will result in the following output:

```console
18:24:03 [INF] [APP] ---------------------------------------------------
18:24:03 [INF] [APP] BLE Sync Cycle v0.60.0 starting...
18:24:03 [INF] [APP] ---------------------------------------------------

Uninstalling the following BLE Sync Cycle v0.60.0 files...

Binary:       /home/richbl/.local/bin/ble-sync-cycle
Desktop file: /home/richbl/.local/share/applications/com.github.richbl.ble-sync-cycle.desktop
Icon:         /home/richbl/.local/share/icons/hicolor/scalable/apps/com.github.richbl.ble-sync-cycle.svg

Uninstallation completed successfully.

18:24:03 [INF] [APP] ---------------------------------------------------
18:24:03 [INF] [APP] BLE Sync Cycle v0.60.0 shutdown complete. Goodbye
18:24:03 [INF] [APP] ---------------------------------------------------
