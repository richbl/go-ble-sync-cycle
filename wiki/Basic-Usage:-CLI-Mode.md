<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->
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
14:40:40 [INF] [APP] ---------------------------------------------------
14:40:40 [INF] [APP] BLE Sync Cycle v0.50.0 starting...
14:40:40 [INF] [APP] ---------------------------------------------------
14:40:40 [DBG] [APP] running in CLI mode
14:40:40 [DBG] [APP] session startup sequence starting...
14:40:40 [DBG] [APP] creating ShutdownManager object (id:0001)...
14:40:40 [DBG] [APP] created ShutdownManager object (id:0001)
14:40:40 [DBG] [APP] ShutdownManager object state stored
14:40:40 [DBG] [APP] initializing controllers...
14:40:40 [DBG] [APP] creating and initializing controllers...
14:40:40 [DBG] [APP] creating new speed controller...
14:40:40 [DBG] [SPD] creating speed controller object (id:0001)...
14:40:40 [DBG] [SPD] created speed controller object (id:0001)
14:40:40 [DBG] [APP] creating new video controller...
14:40:40 [DBG] [VID] creating video controller object (id:0001)...
14:40:40 [INF] [VID] mpv player object created
14:40:40 [DBG] [VID] created video controller object (id:0001)
14:40:40 [DBG] [APP] creating new BLE controller...
14:40:40 [DBG] [BLE] creating BLE controller object (id:0001)...
14:40:40 [DBG] [BLE] created BLE controller object (id:0001)
14:40:40 [DBG] [APP] all controllers created and initialized
14:40:40 [DBG] [APP] controllers initialized OK
14:40:40 [DBG] [APP] establishing connection to BLE peripheral...
14:40:40 [DBG] [BLE] scanning for BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
14:40:40 [DBG] [BLE] BLE peripheral found; stopping scan...
14:40:40 [DBG] [BLE] scan completed
14:40:40 [INF] [BLE] found BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
14:40:40 [DBG] [BLE] connecting to BLE peripheral BD_ADDR=FA:46:1D:77:C8:E1
14:40:40 [INF] [BLE] BLE peripheral connected
14:40:40 [DBG] [BLE] discovering battery service UUID=0000180f-0000-1000-8000-00805f9b34fb
14:40:40 [INF] [BLE] found battery service
14:40:40 [DBG] [BLE] discovering battery characteristic UUID=00002a19-0000-1000-8000-00805f9b34fb
14:40:41 [DBG] [BLE] found battery characteristic UUID=00002a19-0000-1000-8000-00805f9b34fb
14:40:41 [INF] [BLE] BLE sensor battery level: 91%
14:40:41 [DBG] [BLE] discovering CSC service UUID=00001816-0000-1000-8000-00805f9b34fb
14:40:41 [DBG] [BLE] found CSC service UUID=00001816-0000-1000-8000-00805f9b34fb
14:40:41 [DBG] [BLE] discovering CSC characteristic UUID=00002a5b-0000-1000-8000-00805f9b34fb
14:40:41 [DBG] [BLE] found CSC characteristic UUID=00002a5b-0000-1000-8000-00805f9b34fb
14:40:41 [DBG] [APP] BLE peripheral now connected
14:40:41 [DBG] [APP] starting services...
14:40:41 [DBG] [APP] starting BLE service goroutine
14:40:41 [DBG] [APP] starting video service goroutine
14:40:41 [DBG] [APP] BLE and video services started
14:40:41 [DBG] [APP] services started
14:40:41 [DBG] [APP] BLE service starting
14:40:41 [DBG] [BLE] starting the monitoring for BLE sensor notifications...
14:40:41 [DBG] [APP] session startup sequence completed
14:40:41 [DBG] [APP] video service starting
14:40:41 [INF] [VID] starting mpv video playback...
14:40:41 [DBG] [VID] attempting to load file: /home/richbl/Downloads/test_videos/BSC_placeholder_video_14s.mp4
14:40:41 [DBG] [VID] command succeeded, now validating file...
14:40:41 [DBG] [VID] starting mpv video file validation loop
14:40:41 [DBG] [VID] media file successfully loaded: validating dimensions
14:40:41 [DBG] [VID] checkVideoDimensions: starting dimension check
14:40:41 [DBG] [VID] validation successful, draining remaining events
14:40:41 [INF] [VID] video file validation succeeded
14:40:42 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
14:40:42 [DBG] [VID] smoothed sensor speed: 0.00 mph
14:40:42 [DBG] [VID] last playback speed: 0.00 mph
14:40:42 [DBG] [VID] sensor speed delta: 0.00 mph
14:40:42 [DBG] [VID] playback speed update threshold: 0.00 mph
14:40:42 [DBG] [VID] no speed detected, pausing video
14:40:42 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
14:40:42 [DBG] [VID] smoothed sensor speed: 0.00 mph
14:40:42 [DBG] [VID] last playback speed: 0.00 mph
14:40:42 [DBG] [VID] sensor speed delta: 0.00 mph
14:40:42 [DBG] [VID] playback speed update threshold: 0.00 mph
14:40:42 [DBG] [VID] no speed detected, pausing video
14:40:42 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
...
```

In this example (above), the application found the peripheral CSC and battery services and characteristics and is now running in a loop, listening to the BLE peripheral for speed data. The application will also update the playback speed of the media player to match the speed of the sensor. Here, since the video has just begun, its speed is set to 0.0 (paused).

Now let's watch as the speed sensor begins sensing and reporting cycling speed. Note how the smoothing buffer is averaging the speed over the last 5 samples...

```console
...
14:45:07 [DBG] [VID] sensor speed buffer: [14.06 14.35 14.31]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.24 mph
14:45:07 [DBG] [VID] last playback speed: 14.24 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.57x...
14:45:07 [DBG] [VID] sensor speed buffer: [14.06 14.35 14.31]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.24 mph
14:45:07 [DBG] [VID] last playback speed: 14.24 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.57x...
14:45:07 [DBG] [SPD] BLE sensor speed: 13.52 mph
14:45:07 [DBG] [VID] sensor speed buffer: [14.35 14.31 13.52]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.06 mph
14:45:07 [DBG] [VID] last playback speed: 14.24 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.18 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.55x...
14:45:07 [DBG] [VID] sensor speed buffer: [14.35 14.31 13.52]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.06 mph
14:45:07 [DBG] [VID] last playback speed: 14.06 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.55x...
14:45:07 [DBG] [VID] sensor speed buffer: [14.35 14.31 13.52]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.06 mph
14:45:07 [DBG] [VID] last playback speed: 14.06 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.55x...
14:45:07 [DBG] [VID] sensor speed buffer: [14.35 14.31 13.52]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.06 mph
14:45:07 [DBG] [VID] last playback speed: 14.06 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.55x...
14:45:07 [DBG] [VID] sensor speed buffer: [14.35 14.31 13.52]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.06 mph
14:45:07 [DBG] [VID] last playback speed: 14.06 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.55x...
14:45:07 [DBG] [VID] sensor speed buffer: [14.35 14.31 13.52]
14:45:07 [DBG] [VID] smoothed sensor speed: 14.06 mph
14:45:07 [DBG] [VID] last playback speed: 14.06 mph
14:45:07 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:07 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:07 [DBG] [VID] updating video playback speed to 1.55x...
...
```

In this last example, **BLE Sync Cycle** is coordinating with both the BLE peripheral (the speed sensor) and the video player, updating the video player to match the speed of the sensor.

Finally, let's watch when the user stops the **BLE Sync Cycle** application...

**To quit the application, press `Ctrl+C`.**

```console
...
14:45:08 [DBG] [VID] updating video playback speed to 1.55x...
14:45:08 [DBG] [SPD] BLE sensor speed: 13.41 mph
14:45:08 [DBG] [VID] sensor speed buffer: [14.31 13.52 13.41]
14:45:08 [DBG] [VID] smoothed sensor speed: 13.75 mph
14:45:08 [DBG] [VID] last playback speed: 14.06 mph
14:45:08 [DBG] [VID] sensor speed delta: 0.31 mph
14:45:08 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:08 [DBG] [VID] updating video playback speed to 1.51x...
14:45:08 [DBG] [VID] sensor speed buffer: [14.31 13.52 13.41]
14:45:08 [DBG] [VID] smoothed sensor speed: 13.75 mph
14:45:08 [DBG] [VID] last playback speed: 13.75 mph
14:45:08 [DBG] [VID] sensor speed delta: 0.00 mph
14:45:08 [DBG] [VID] playback speed update threshold: 0.00 mph
14:45:08 [DBG] [VID] updating video playback speed to 1.51x...
14:45:08 [INF] [APP] shutdown request detected, shutting down now...
14:45:08 [DBG] [APP] shutting down ShutdownManager object (id:0001)...
14:45:08 [DBG] [VID] interrupt detected, stopping mpv video playback...
14:45:08 [DBG] [VID] terminating video controller object (id:0001)...
14:45:08 [DBG] [VID] starting player termination
14:45:08 [DBG] [BLE] interrupt detected, stopping the monitoring for BLE sensor notifications...
14:45:08 [DBG] [APP] shutting down ShutdownManager object (id:0001)...
14:45:08 [DBG] [VID] call to terminate mpv completed successfully
14:45:08 [DBG] [VID] destroyed MPV handle: C resources released
14:45:08 [DBG] [VID] destroyed video controller object (id:0001)
14:45:08 [DBG] [APP] ShutdownManager (id:0001) services stopped
14:45:08 [DBG] [APP] ShutdownManager (id:0001) services stopped
14:45:08 [DBG] [APP] ShutdownManager object (id:0001) shutdown complete
14:45:08 [DBG] [APP] ShutdownManager object (id:0001) shutdown complete
14:45:08 [INF] [APP] ---------------------------------------------------
14:45:08 [INF] [APP] BLE Sync Cycle v0.50.0 shutdown complete. Goodbye
14:45:08 [INF] [APP] ---------------------------------------------------
```
