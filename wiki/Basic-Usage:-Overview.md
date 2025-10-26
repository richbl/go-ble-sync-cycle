<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

![bsc_example](https://github.com/user-attachments/assets/0eb908da-6d22-4e8a-a9e8-5702a01f3fe0)

At a high level, **BLE Sync Cycle** follows the steps below:

1. Scans for your BLE cycling sensor
2. Connects to the sensor and starts receiving speed data
3. Launches a media player for video playback
4. Automatically adjusts video speed based on your cycling speed
5. Displays real-time cycling statistics via the on-screen display (OSD)
6. Gracefully shuts down on user interrupt (Ctrl+C) or end of video playback

### Running the Application

To run the application, you need to first make sure that your Bluetooth devices are enabled and in range before running this command. On a computer or similar, you should have your Bluetooth radio turned on. On a BLE sensor, you typically "wake it up" by moving or shaking the device (i.e., spinning the bicycle wheel).

To run **BLE Sync Cycle**, execute the following command:

```console
./ble-sync-cycle
```

If the application hasn't yet been built using the `go build` command, please refer to the [Building the Application](#building-the-application) section above.

> **IMPORTANT:** Be sure the `config.toml` is located in the current working directory (where you ran the `ble-sync-cycle` command), or see the next section on how to override where the application looks for a configuration file.
