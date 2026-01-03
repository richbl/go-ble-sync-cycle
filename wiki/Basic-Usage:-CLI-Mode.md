<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

### CLI Mode: Running the Application from the Command Line

To run the application, you need to first make sure that your Bluetooth devices are enabled and in range before running this command. On a computer or similar, you should have your Bluetooth radio turned on. On a BLE sensor, you typically "wake it up" by moving or shaking the device (i.e., spinning the bicycle wheel).

To run **BLE Sync Cycle**, execute the following command:

```console
./ble-sync-cycle --no-gui
```

Note the use of the `--no-gui` (or `-n`) flag. This flag tells **BLE Sync Cycle** to run in CLI mode, rather than GUI mode. To learn more about running **BLE Sync Cycle** in GUI mode, see [Basic Usage: GUI Mode](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-GUI-Mode).

Be sure the default project `config.toml` is located in the current working directory (where you ran the `ble-sync-cycle` command), or see [Using the Command Line Options](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-Using-the-Command-Line-Options) to learn how to override where **BLE Sync Cycle** looks for a configuration file.

### Monitoring the Session Log

The session log is output to the console when **BLE Sync Cycle** is running in CLI mode. This is the primary mechanism for reporting application information to the user.

After starting `ble-sync-cycle`, you'll see the following output.

Note that the output below was generated when `logging_level` is set to `debug` in the `config.toml` file. This means that all log message types (debug, info, warn, error, and fatal) will be displayed.

```console
13:14:24 [INF] [APP] BLE Sync Cycle v0.13.0 starting...
13:14:24 [INF] [APP] running in CLI mode
13:14:24 [DBG] [APP] logging level changed to DEBUG
13:14:24 [DBG] [APP] set PendingStart=true, state=Connecting
13:14:24 [DBG] [APP] creating ShutdownManager
13:14:24 [DBG] [APP] shutdownMgr stored
13:14:24 [DBG] [APP] initializing controllers...
13:14:24 [DBG] [APP] creating speed controller...
13:14:24 [DBG] [APP] creating video controller...
13:14:24 [DBG] [APP] creating BLE controller...
13:14:24 [INF] [BLE] created new BLE central controller
13:14:24 [DBG] [APP] controllers initialized OK
13:14:24 [DBG] [APP] connecting BLE...
13:14:24 [DBG] [BLE] scanning for BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
13:14:54 [ERR] [APP] BLE connect failed: BLE scan failed: scanning time limit reached (30s)
13:14:54 [FTL] [APP] BLE connection failed: BLE scan failed: scanning time limit reached (30s)
13:14:54 [INF] [APP] BLE Sync Cycle v0.13.0 shutdown complete. Goodbye
```

In this first example (above), while the application was able to find the BLE peripheral, it failed to discover the CSC services and characteristics before timing out.

Depending on the BLE peripheral, it may take some time before a BLE peripheral advertises both its services and characteristics. If the peripheral is not responding, you may need to increase the timeout in the `config.toml` file. In most cases however, rerunning the application will resolve the issue.

Let's look at a second example, where the BLE peripheral is responding and the CSC services and characteristics are found...

```console
13:29:15 [INF] [APP] BLE Sync Cycle v0.13.0 starting...
13:29:15 [INF] [APP] running in CLI mode
13:29:15 [DBG] [APP] logging level changed to DEBUG
13:29:15 [DBG] [APP] set PendingStart=true, state=Connecting
13:29:15 [DBG] [APP] creating ShutdownManager
13:29:15 [DBG] [APP] shutdownMgr stored
13:29:15 [DBG] [APP] initializing controllers...
13:29:15 [DBG] [APP] creating speed controller...
13:29:15 [DBG] [APP] creating video controller...
13:29:15 [DBG] [APP] creating BLE controller...
13:29:15 [INF] [BLE] created new BLE central controller
13:29:15 [DBG] [APP] controllers initialized OK
13:29:15 [DBG] [APP] connecting BLE...
13:29:15 [DBG] [BLE] scanning for BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
13:29:15 [INF] [BLE] found BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
13:29:15 [DBG] [BLE] connecting to BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
13:29:15 [INF] [BLE] BLE peripheral device connected
13:29:15 [DBG] [BLE] discovering battery service UUID=0000180f-0000-1000-8000-00805f9b34fb
13:29:15 [INF] [BLE] found battery service
13:29:15 [DBG] [BLE] discovering battery characteristic UUID=00002a19-0000-1000-8000-00805f9b34fb
13:29:16 [INF] [BLE] found battery characteristic UUID=00002a19-0000-1000-8000-00805f9b34fb
13:29:16 [INF] [BLE] BLE sensor battery level: 99%
13:29:16 [DBG] [BLE] discovering CSC service UUID=00001816-0000-1000-8000-00805f9b34fb
13:29:16 [INF] [BLE] found CSC service UUID=00001816-0000-1000-8000-00805f9b34fb
13:29:16 [DBG] [BLE] discovering CSC characteristic UUID=00002a5b-0000-1000-8000-00805f9b34fb
13:29:16 [INF] [BLE] found CSC characteristic UUID=00002a5b-0000-1000-8000-00805f9b34fb
13:29:16 [DBG] [APP] BLE connected OK
13:29:16 [DBG] [APP] set state=Running, PendingStart=false
13:29:16 [DBG] [APP] starting services...
13:29:16 [DBG] [APP] services started
13:29:16 [INF] [VID] starting mpv video playback...
13:29:16 [INF] [BLE] starting the monitoring for BLE sensor notifications...
13:29:16 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
13:29:16 [DBG] [VID] smoothed sensor speed: 0.00 mph
13:29:16 [DBG] [VID] last playback speed: 0.00 mph
13:29:16 [DBG] [VID] sensor speed delta: 0.00 mph
13:29:16 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:16 [DBG] [VID] no speed detected, pausing video
...
```

In this example (above), the application found the peripheral CSC and battery services and characteristics and is now running in a loop, listening to the BLE peripheral for speed data. The application will also update the playback speed of the media player to match the speed of the sensor. Here, since the video has just begun, its speed is set to 0.0 (paused).

Now let's watch as the speed sensor begins sensing and reporting cycling speed. Note how the smoothing buffer is averaging the speed over the last 5 samples...

```console
...
13:29:18 [DBG] [SPD] BLE sensor speed: 16.56 mph
13:29:18 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 16.56]
13:29:18 [DBG] [VID] smoothed sensor speed: 3.31 mph
13:29:18 [DBG] [VID] last playback speed: 0.00 mph
13:29:18 [DBG] [VID] sensor speed delta: 3.31 mph
13:29:18 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:18 [DBG] [VID] updating video playback speed to 0.26x...
13:29:18 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 16.56]
13:29:18 [DBG] [VID] smoothed sensor speed: 3.31 mph
13:29:18 [DBG] [VID] last playback speed: 3.31 mph
13:29:18 [DBG] [VID] sensor speed delta: 0.00 mph
13:29:18 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:18 [DBG] [VID] updating video playback speed to 0.26x...
13:29:18 [DBG] [SPD] BLE sensor speed: 16.13 mph
13:29:18 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 16.56 16.13]
13:29:18 [DBG] [VID] smoothed sensor speed: 6.54 mph
13:29:18 [DBG] [VID] last playback speed: 3.31 mph
13:29:18 [DBG] [VID] sensor speed delta: 3.23 mph
13:29:18 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:18 [DBG] [VID] updating video playback speed to 0.52x...
13:29:19 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 16.56 16.13]
13:29:19 [DBG] [VID] smoothed sensor speed: 6.54 mph
13:29:19 [DBG] [VID] last playback speed: 6.54 mph
13:29:19 [DBG] [VID] sensor speed delta: 0.00 mph
13:29:19 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:19 [DBG] [VID] updating video playback speed to 0.52x...
13:29:19 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 16.56 16.13]
13:29:19 [DBG] [VID] smoothed sensor speed: 6.54 mph
13:29:19 [DBG] [VID] last playback speed: 6.54 mph
13:29:19 [DBG] [VID] sensor speed delta: 0.00 mph
13:29:19 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:19 [DBG] [VID] updating video playback speed to 0.52x...
13:29:19 [DBG] [SPD] BLE sensor speed: 15.47 mph
13:29:19 [DBG] [VID] sensor speed buffer: [0.00 0.00 16.56 16.13 15.47]
13:29:19 [DBG] [VID] smoothed sensor speed: 9.63 mph
13:29:19 [DBG] [VID] last playback speed: 6.54 mph
13:29:19 [DBG] [VID] sensor speed delta: 3.09 mph
13:29:19 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:19 [DBG] [VID] updating video playback speed to 0.77x...
```

In this last example, **BLE Sync Cycle** is coordinating with both the BLE peripheral (the speed sensor) and the video player, updating the video player to match the speed of the sensor.

Finally, let's watch when the user stops the **BLE Sync Cycle** application...

**To quit the application, press `Ctrl+C`.**

```console
...
13:29:21 [DBG] [VID] sensor speed buffer: [15.47 15.52 15.75 15.80 15.43]
13:29:21 [DBG] [VID] smoothed sensor speed: 15.59 mph
13:29:21 [DBG] [VID] last playback speed: 15.59 mph
13:29:21 [DBG] [VID] sensor speed delta: 0.00 mph
13:29:21 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:21 [DBG] [VID] updating video playback speed to 1.25x...
13:29:22 [DBG] [SPD] BLE sensor speed: 15.19 mph
13:29:22 [DBG] [VID] sensor speed buffer: [15.52 15.75 15.80 15.43 15.19]
13:29:22 [DBG] [VID] smoothed sensor speed: 15.54 mph
13:29:22 [DBG] [VID] last playback speed: 15.59 mph
13:29:22 [DBG] [VID] sensor speed delta: 0.06 mph
13:29:22 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:22 [DBG] [VID] updating video playback speed to 1.24x...
13:29:22 [DBG] [VID] sensor speed buffer: [15.52 15.75 15.80 15.43 15.19]
13:29:22 [DBG] [VID] smoothed sensor speed: 15.54 mph
13:29:22 [DBG] [VID] last playback speed: 15.54 mph
13:29:22 [DBG] [VID] sensor speed delta: 0.00 mph
13:29:22 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:22 [DBG] [VID] updating video playback speed to 1.24x...
13:29:22 [DBG] [SPD] BLE sensor speed: 14.87 mph
13:29:22 [DBG] [SPD] BLE sensor speed: 14.96 mph
13:29:22 [DBG] [VID] sensor speed buffer: [15.80 15.43 15.19 14.87 14.96]
13:29:22 [DBG] [VID] smoothed sensor speed: 15.25 mph
13:29:22 [DBG] [VID] last playback speed: 15.54 mph
13:29:22 [DBG] [VID] sensor speed delta: 0.29 mph
13:29:22 [DBG] [VID] playback speed update threshold: 0.25 mph
13:29:22 [DBG] [VID] updating video playback speed to 1.22x...
13:29:22 [INF] [BLE] interrupt detected, stopping the monitoring for BLE sensor notifications...
13:29:22 [INF] [VID] interrupt detected, stopping mpv video playback...
13:29:22 [INF] [APP] BLE Sync Cycle v0.13.0 shutdown complete. Goodbye
```
