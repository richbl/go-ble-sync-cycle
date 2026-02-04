<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

**BLE Sync Cycle** can be run in two application modes: GUI or CLI.

### Running **BLE Sync Cycle** in GUI Mode

In GUI (graphical user interface) mode, the application is started and run like any other graphical application. Some benefits of running in GUI mode include:

- Easy to start
- Easy to visually track video progression and real-time cycling metrics during a cycling session
- Real-time logging is available directly within the application interface
- Easy to manage and edit multiple BSC sessions

> Note that in order to run **BLE Sync Cycle** in GUI mode, the application must be installed as a desktop application (see [Installing the Application](https://github.com/richbl/go-ble-sync-cycle/wiki/Installation:-Installing-the-Application)) for details.

### Running **BLE Sync Cycle** in CLI Mode

In CLI (command line interface) mode, the application is started from the command line within a terminal. Some benefits of running in CLI mode include:

- Easy to run the application from anywhere within a terminal
- The application can be configured with optional command-line configuration flags
- Real-time logging is available (outputs to `stdout`)
- Easy to shutdown the application (CTRL+C, which sends a `SIGINT` signal to the application)

> Running **BLE Sync Cycle** in CLI mode does not require any additional installation beyond building the application (see [Building the Application](https://github.com/richbl/go-ble-sync-cycle/wiki/Installation:-Building-the-Application)).
