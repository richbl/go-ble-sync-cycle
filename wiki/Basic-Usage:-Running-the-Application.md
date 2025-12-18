<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

After starting `ble-sync-cycle`, you should see the following output. Note that the output below was generated when `logging_level` was set to `debug` in the `config.toml` file.

```console
2025/10/26 13:11:27 ----- ----- Starting BLE Sync Cycle v0.13.0
2025/10/26 13:11:27 [INF] [BLE] created new BLE central controller
2025/10/26 13:11:27 [DBG] [BLE] scanning for BLE peripheral BD_ADDR FA:46:1D:77:C8:E1
2025/10/26 13:11:27 [INF] [BLE] found BLE peripheral FA:46:1D:77:C8:E1
2025/10/26 13:11:27 [DBG] [BLE] connecting to BLE peripheral FA:46:1D:77:C8:E1
2025/10/26 13:11:27 [INF] [BLE] BLE peripheral device connected
2025/10/26 13:11:27 [DBG] [BLE] discovering CSC service 00001816-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:32 [FTL] [BLE] failed to acquire BLE services: scanning time limit reached (30s)
2025/10/26 13:11:32 ----- ----- BLE Sync Cycle v0.13.0 shutdown complete. Goodbye
```

In this first example (above), while the application was able to find the BLE peripheral, it failed to discover the CSC services and characteristics before timing out. Depending on the BLE peripheral, it may take some time before a BLE peripheral advertises both its services and characteristics. If the peripheral is not responding, you may need to increase the timeout in the `config.toml` file. In most cases however, rerunning the application will resolve the issue.

```console
2025/10/26 13:11:37 ----- ----- Starting BLE Sync Cycle v0.13.0
2025/10/26 13:11:38 [INF] [BLE] created new BLE central controller
2025/10/26 13:11:38 [DBG] [BLE] scanning for BLE peripheral BD_ADDR FA:46:1D:77:C8:E1
2025/10/26 13:11:38 [INF] [BLE] found BLE peripheral FA:46:1D:77:C8:E1
2025/10/26 13:11:38 [DBG] [BLE] connecting to BLE peripheral FA:46:1D:77:C8:E1
2025/10/26 13:11:38 [INF] [BLE] BLE peripheral device connected
2025/10/26 13:11:38 [DBG] [BLE] discovering CSC service 00001816-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [INF] [BLE] found CSC service 00001816-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [DBG] [BLE] discovering CSC characteristic 00002a5b-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [INF] [BLE] found CSC characteristic 00002a5b-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [DBG] [BLE] discovering battery service 0000180f-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [INF] [BLE] found battery service 0000180f-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [DBG] [BLE] discovering battery characteristic 00002a19-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:40 [INF] [BLE] found battery characteristic 00002a19-0000-1000-8000-00805f9b34fb
2025/10/26 13:11:45 [INF] [BLE] BLE sensor battery level: 100%
2025/10/26 13:11:45 [INF] [VID] starting mpv video playback...
2025/10/26 13:11:45 [INF] [BLE] starting the monitoring for BLE sensor notifications...
2025/10/26 13:11:45 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
2025/10/26 13:11:45 [DBG] [VID] smoothed sensor speed: 0.00 mph
2025/10/26 13:11:45 [DBG] [VID] last playback speed: 0.00 mph
2025/10/26 13:11:45 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:11:45 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:11:45 [DBG] [VID] no speed detected, pausing video
...
```

In the next example (above), the application found the peripheral CSC and battery services and characteristics and is now running in a loop, listening to the BLE peripheral for speed data. The application will also update the playback speed of the media player to match the speed of the sensor. Here, since the video has just begun, its speed is set to 0.0 (paused).

```console
...
2025/10/26 13:12:20 [DBG] [SPD] BLE sensor speed: 10.87 mph
2025/10/26 13:12:20 [DBG] [VID] sensor speed buffer: [11.64 11.71 11.51 11.09 10.87]
2025/10/26 13:12:20 [DBG] [VID] smoothed sensor speed: 11.36 mph
2025/10/26 13:12:20 [DBG] [VID] last playback speed: 11.89 mph
2025/10/26 13:12:20 [DBG] [VID] sensor speed delta: 0.52 mph
2025/10/26 13:12:20 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:12:20 [DBG] [VID] updating video playback speed to 0.91x
2025/10/26 13:12:20 [DBG] [VID] sensor speed buffer: [11.64 11.71 11.51 11.09 10.87]
2025/10/26 13:12:20 [DBG] [VID] smoothed sensor speed: 11.36 mph
2025/10/26 13:12:20 [DBG] [VID] last playback speed: 11.36 mph
2025/10/26 13:12:20 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:12:20 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:12:20 [DBG] [VID] updating video playback speed to 0.91x
2025/10/26 13:12:20 [DBG] [SPD] BLE sensor speed: 11.22 mph
2025/10/26 13:12:20 [DBG] [VID] sensor speed buffer: [11.71 11.51 11.09 10.87 11.22]
2025/10/26 13:12:20 [DBG] [VID] smoothed sensor speed: 11.28 mph
2025/10/26 13:12:20 [DBG] [VID] last playback speed: 11.36 mph
2025/10/26 13:12:20 [DBG] [VID] sensor speed delta: 0.08 mph
2025/10/26 13:12:20 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:12:20 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:12:21 [DBG] [VID] sensor speed buffer: [11.71 11.51 11.09 10.87 11.22]
2025/10/26 13:12:21 [DBG] [VID] smoothed sensor speed: 11.28 mph
2025/10/26 13:12:21 [DBG] [VID] last playback speed: 11.28 mph
2025/10/26 13:12:21 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:12:21 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:12:21 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:12:21 [DBG] [VID] sensor speed buffer: [11.71 11.51 11.09 10.87 11.22]
2025/10/26 13:12:21 [DBG] [VID] smoothed sensor speed: 11.28 mph
2025/10/26 13:12:21 [DBG] [VID] last playback speed: 11.28 mph
2025/10/26 13:12:21 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:12:21 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:12:21 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:12:21 [DBG] [SPD] BLE sensor speed: 11.4 mph
2025/10/26 13:12:21 [DBG] [VID] sensor speed buffer: [11.51 11.09 10.87 11.22 11.40]

```

In this last example, **BLE Sync Cycle** is coordinating with both the BLE peripheral (the speed sensor) and the video player, updating the video player to match the speed of the sensor.

**To quit the application, press `Ctrl+C`.**

```console
...
2025/10/26 13:18:21 [DBG] [VID] sensor speed buffer: [11.71 11.51 11.09 10.87 11.22]
2025/10/26 13:18:21 [DBG] [VID] smoothed sensor speed: 11.28 mph
2025/10/26 13:18:21 [DBG] [VID] last playback speed: 11.28 mph
2025/10/26 13:18:21 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:18:21 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:18:21 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:18:21 [DBG] [SPD] BLE sensor speed: 11.4 mph
2025/10/26 13:18:21 [DBG] [VID] sensor speed buffer: [11.51 11.09 10.87 11.22 11.40]
2025/10/26 13:18:21 [DBG] [VID] smoothed sensor speed: 11.22 mph
2025/10/26 13:18:21 [DBG] [VID] last playback speed: 11.28 mph
2025/10/26 13:18:21 [DBG] [VID] sensor speed delta: 0.06 mph
2025/10/26 13:18:21 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:18:21 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:18:21 [DBG] [VID] sensor speed buffer: [11.51 11.09 10.87 11.22 11.40]
2025/10/26 13:18:21 [DBG] [VID] smoothed sensor speed: 11.22 mph
2025/10/26 13:18:21 [DBG] [VID] last playback speed: 11.22 mph
2025/10/26 13:18:21 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:18:21 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:18:21 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:18:22 [DBG] [VID] sensor speed buffer: [11.51 11.09 10.87 11.22 11.40]
2025/10/26 13:18:22 [DBG] [VID] smoothed sensor speed: 11.22 mph
2025/10/26 13:18:22 [DBG] [VID] last playback speed: 11.22 mph
2025/10/26 13:18:22 [DBG] [VID] sensor speed delta: 0.00 mph
2025/10/26 13:18:22 [DBG] [VID] playback speed update threshold: 0.25 mph
2025/10/26 13:18:22 [DBG] [VID] updating video playback speed to 0.90x
2025/10/26 13:18:22 [INF] [VID] interrupt detected, stopping MPV video playback...
2025/10/26 13:18:22 [INF] [BLE] interrupt detected, stopping the monitoring for BLE sensor notifications...
2025/10/26 13:18:22 ----- ----- BLE Sync Cycle v0.13.0 shutdown complete. Goodbye
```
