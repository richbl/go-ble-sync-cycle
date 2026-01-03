<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

**BLE Sync Cycle** can be run in two application modes: GUI or CLI.

### Running **BLE Sync Cycle** in GUI Mode

In GUI (graphical user interface) mode, the application is started and run like any other graphical application. Some benefits of running in GUI mode include:

- Easy to start
- Easy to visually track video progression and real-time cycling metrics during a cycling session
- Real-time logging is available directly within the application interface
- Easy to manage and edit multiple BSC sessions

> Note that in order to run **BLE Sync Cycle** in GUI mode, the application must be installed as a desktop application (see [Installing the Application](#installing-the-application)) for details.

### Running **BLE Sync Cycle** in CLI Mode

In CLI (command line interface) mode, the application is started from the command line within a terminal. Some benefits of running in CLI mode include:

- Easy to run the application from anywhere within a terminal
- The application can be configured with optional command-line configuration flags
- Real-time logging is available (outputs to `stdout`)
- Easy to shutdown the application (CTRL+C, which sends a `SIGINT` signal to the application)

> Running **BLE Sync Cycle** in CLI mode does not require any additional installation beyond building the application (see [Building the Application](#building-the-application)).
