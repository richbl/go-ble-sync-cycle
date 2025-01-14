# BLE Sync Cycle

![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/richbl/go-ble-sync-cycle?include_prereleases)
[![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-ble-sync-cycle)](https://goreportcard.com/report/github.com/richbl/go-ble-sync-cycle) [![codebeat badge](https://codebeat.co/badges/7d948b80-136b-41be-9afd-1604f7dce6fa)](https://codebeat.co/projects/github-com-richbl-go-ble-sync-cycle-dev) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/595889e53f25475da18dea64b5a60419)](https://app.codacy.com/gh/richbl/go-ble-sync-cycle/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=richbl_go-ble-sync-cycle&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=richbl_go-ble-sync-cycle)

## Overview

**BLE Sync Cycle** is a Go application designed to synchronize video playback with real-time cycling data from Bluetooth Low Energy (BLE) devices, such as cycling speed and cadence (CSC) sensors. This integration provides users with a more immersive indoor cycling experience by matching video playback with their actual cycling pace, making it a valuable option when outdoor cycling isn't feasible.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><img src="[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9)" width=700></picture>
</p>

## Features

- Real-time synchronization between cycling speed and video playback
- Support for compliant Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensors
- TOML-based configuration for easy application customization that includes:
    - BLE sensor identification (UUID)
    - Bluetooth device scanning timeout
    - Wheel circumference, for accurate speed conversion
    - Support for different speed units: miles per hour (mph), kilometers per hour (kph)
    - Speed smoothing option for natural and seamless video playback
    - Choice of video file for playback
    - Various display options for video playback, including:
        - The display of sensor speed, video playback speed and playback time remaining via on-screen display (OSD)
        - Video window scaling (full screen, half screen, etc.)
        - OSD font size
- Simple command-line interface provides real-time component feedback
- Configurable logging levels (debug, info, warn, error) to manage the information displayed during application execution
- Graceful handling of connection interrupts and system signals ensures all components shut down cleanly

## Rationale

This project was developed to address a specific need: **how can I remain engaged in cycling when the weather outside is less than ideal?**

While there are several existing solutions that allow for "virtual" indoor cycling, such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), these typically require the purchase of specialized training equipment (often preventing the use of your own bike), a subscription to compatible online virtual cycling services, and a reliable broadband Internet connection. My needs, however, are different:

- I want to train using my own bicycle. Since I prefer riding recumbents, it wouldn’t make sense for me to "train" on a traditional upright trainer
- I need a solution that can operate independently without requiring an Internet connection, as I live in a rural area of the Pacific Northwest where both electricity and Internet access can be unreliable at best

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle our regular loss of Internet service

- I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with (I suspect it's my nature as an engineer)

Since I already use a mechanical (no electronics) portable bicycle trainer while riding indoors, it made sense for me to find a way to pair my existing Bluetooth cycling sensors with a local computer which could then drive some kind of interesting feedback while cycling. This project was created to fit that need.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6"><img src="[[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6)]([https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6))" width=700></picture>
</p>

## Requirements

### Hardware Components

- A bicycle set up for indoor riding
- A Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensor, configured for speed
- A computer that supports Bluetooth (4.0+), preferably with a big screen display to watch video playback

For my own indoor cycling configuration, I use a Performance Travel Trac 3 trainer. The BLE sensor used is a [Magene S3+ Speed/Cadence Dual Mode Sensor](https://www.magene.com/en/sensors/59-s3-speed-cadence-dual-mode-sensor.html) configured for speed, though any BLE-compliant sensor should work.

> For an overview of Bluetooth BLE, refer to the excellent article ["Introduction to Bluetooth Low Energy"](https://learn.adafruit.com/introduction-to-bluetooth-low-energy/introduction) by Kevin Townsend.

### Software Components

- The open source, cross-platform [mpv media player](https://mpv.io/), installed (e.g., `sudo apt-get install mpv`) and operational
- The `libmpv2` library, installed (e.g., `sudo apt-get install libmpv2`)
- In order to compile the executable for this project, an operational [Go language](https://go.dev/) environment is required (this release was developed using Go 1.23.2). Once the **BLE Sync Cycle** application is compiled into an executable, it can be run without the dependencies on the Go language environment
- A local video file for playback using mpv, preferably a first-person view cycling video. Check out [YouTube with this query: "first person cycling"](https://www.youtube.com/results?search_query=first+person+cycling) for some ideas

While **BLE Sync Cycle** has been written and tested using Ubuntu 24.04 (LTS) on an Intel processor (amd64), it should work across any recent comparable Unix-like platform and architecture.

## Installation

### Install Application Dependencies

**BLE Sync Cycle** currently relies on the mpv media player for video playback (support for additional media players will be implemented in a future release). In order for the application to function, the mpv media player library must first be installed.

1. Install the `libmpv2` library:

    ```console
    sudo apt-get install libmpv2
    ```

### Building the Application

1. Clone the repository:

    ```console
    git clone https://github.com/richbl/go-ble-sync-cycle
    cd go-ble-sync-cycle
    ```

2. Install Go package dependencies:

    ```console
    go mod download
    ```

3. Build the application:

    ```console
    go build -o ble-sync-cycle cmd/*
    ```

The resulting `build` command will create the`ble-sync-cycle` executable in the current directory.

### Editing the TOML File

Edit the `config.toml` file found in the `internal/configuration` directory. The default file (with a different sensor UUID) is shown below:

```toml
# BLE Sync Cycle TOML configuration
 # 0.8.1

  [app]
    logging_level = "debug" # Log messages to see during execution: "debug", "info", "warn", "error"
                            # where "debug" is the most verbose and "error" is least verbose

  [ble]
    sensor_uuid = "F1:42:D8:DE:35:16" # UUID of BLE peripheral device
    scan_timeout_secs = 30            # Seconds to wait for peripheral response before generating error

  [speed]
    smoothing_window = 5          # Number of speed look-backs to use for generating a moving average
    speed_threshold = 0.25        # Minimum speed change to trigger video speed update
    wheel_circumference_mm = 1932 # Wheel circumference in millimeters
    speed_units = "mph"           # "km/h" or "mph"

  [video]
    file_path = "cycling_test.mp4" # Path to the video file to play
    window_scale_factor = 1.0      # Scale factor for the video window (1.0 = full screen)
    update_interval_sec = 0.25     # Seconds (>0.0) to wait between video player updates
    speed_multiplier = 0.6         # Multiplier that translates sensor speed to video playback speed
                                  # (0.0 = stopped, 1.0 = normal speed)
    [video.OSD]
      font_size = 40                # Font size for on-screen display (OSD)
      display_cycle_speed = true    # Display cycle speed on the on-screen display (true/false)
      display_playback_speed = true # Display video playback speed on the on-screen display (true/false)
      display_time_remaining = true # Display time remaining on the on-screen display (true/false)

```

An explanation of the various sections of the `config.toml` file is provided below:

#### The `[app]` Section

The `[app]` section is used for configuration of the **BLE Sync Cycle** application itself. It includes the following parameter:

- `logging_level`: The logging level to use, which displays messages to the console as the application executes. This can be "debug", "info", "warn", or "error", where "debug" is the most verbose and "error" is least verbose.

#### The `[ble]` Section

The `[ble]` section configures your computer (referred to as the BLE central controller) to scan for and query the BLE speed sensor (referred to as the BLE peripheral). It includes the following parameters:

- `sensor_uuid`: The UUID of the BLE peripheral device (e.g., sensor) to connect with and monitor for speed data
- `scan_timeout_secs`: The number of seconds to wait for a BLE peripheral response before generating an error. Some BLE devices can take a while to respond, so adjust this value accordingly.

> To find the UUID of your BLE peripheral device, you'll need to connect to it from your computer (or any device with Bluetooth connectivity). From Ubuntu (or any other Linux distribution), you can use [the `bluetoothctl` command](https://www.mankier.com/1/bluetoothctl#). BLE peripheral device UUIDs are typically in the form of "11:22:33:44:55:66."

#### The `[speed]` Section

The `[speed]` section defines the configuration for the speed controller component. The speed controller takes raw BLE CSC speed data (a rate of discrete device events per time cycle) and converts it speed (either km/h or mph, depending on `speed_units`). It includes the following parameters:

- `smoothing_window`: The number of look-backs (or buffered speed measurements) to use for generating a moving average for the speed value
- `speed_threshold`: The minimum speed change to trigger video speed updates
- `wheel_circumference_mm`: The wheel circumference in millimeters, important in order to accurately convert raw sensor values to actual speed (distance traveled per unit time)
- `speed_units`: The speed units to use (either "km/h" or "mph")

> The smoothing window is a simple ring buffer that stores the last (n) speed measurements, meaning that it will create a moving average for the speed value. This helps to smooth out the speed data and provide a more natural video playback experience.

#### The `[video]` Section

The `[video]` section defines the configuration for the MPV video player component. It includes the following parameters:

- `file_path`: The path to the video file to play. The video format must be supported by MPV (e.g., MP4, webm, etc.)
- `window_scale_factor`: A scaling factor for the video window, where 1.0 is full screen. This value can be useful when debugging or when running the video player in a non-maximized window is useful (e.g., 0.5 = half screen)
- `update_interval_sec`: The number of seconds (>0.0) to wait between video player updates.

> The `speed_multiplier` parameter is used to control the relative playback speed of the video. Usually, a value of 1.0 is used, as this is the default value (normal playback speed). However, since it's typically unknown what the speed of the bicycle rider in the video is during "normal speed" playback, it's recommended to experiment with different values to find a good balance between  video playback speed and real-world cycling experience.

#### The `[video.OSD]` Section

- `font_size`: The font size for the on-screen display (OSD)
- `display_cycle_speed`: A boolean value that indicates whether to display the cycle sensor speed on the on-screen display (OSD)
- `display_playback_speed`: A boolean value that indicates whether to display the video playback speed on the on-screen display (OSD)
- `display_time_remaining`: A boolean value that indicates whether to display the time remaining (using the format HH:MM:SS) on the on-screen display (OSD)

## Basic Usage

At a high level, **BLE Sync Cycle** will perform the following:

1. Scan for your BLE cycling sensor
2. Connect to the sensor and start receiving speed data
3. Launch MPV video playback
4. Automatically adjust video speed based on your cycling speed
5. Gracefully shutdown on interrupt (Ctrl+C)

To run the application, you need to first make sure that your Bluetooth devices are enabled and in range before running this command. On a computer or similar, you should have your Bluetooth radio turned on. On a BLE sensor, you typically "wake it up" by moving or shaking the device (i.e., spinning the bicycle wheel).

To run **BLE Sync Cycle**, execute the following command:

```console
./ble-sync-cycle
```

If the application hasn't yet been built using the `go build` command, please refer to the [Building the Application](#building-the-application) section above.

> Be sure the `config.toml` is located in the current working directory (where you ran the `ble-sync-cycle` command), or see the section below on how to override where the application looks for the `config.toml` file.

### Overriding the Configuration File

When **BLE Sync Cycle** is first started, it looks for a configuration file called `config.toml` in the current working directory. If you want  **BLE Sync Cycle** to look in a different location--you could use different configuration files for different cycling sessions, different bicycle configurations, different sensors, etc.--you can specify the path to the file on the command line using the `--config` command line option:

```console
./ble-sync-cycle --config /path/to/config.toml
```

Or, you can also use the `-c` command line option for the same behavior:

```console
./ble-sync-cycle -c /path/to/config.toml
```

### Running the Application

At this point, you should see the following output:

```console
 2024/12/16 15:08:56 ----- ----- Starting BLE Sync Cycle 0.8.1
  2024/12/16 15:08:56 [INF] [BLE] created new BLE central controller
  2024/12/16 15:08:56 [INF] [BLE] now scanning the ether for BLE peripheral UUID of F1:42:D8:DE:35:16...
  2024/12/16 15:08:58 [DBG] [BLE] found BLE peripheral F1:42:D8:DE:35:16
  2024/12/16 15:08:58 [DBG] [BLE] connecting to BLE peripheral device F1:42:D8:DE:35:16
  2024/12/16 15:09:00 [INF] [BLE] BLE peripheral device connected
  2024/12/16 15:09:00 [DBG] [BLE] discovering CSC services 00001816-0000-1000-8000-00805f9b34fb
  2024/12/16 15:09:10 [ERR] [BLE] CSC services discovery failed: timeout on DiscoverServices
  2024/12/16 15:09:10 [ERR] [BLE] BLE peripheral scan failed: timeout on DiscoverServices
  2024/12/16 15:09:10 ----- ----- BLE Sync Cycle 0.8.1 shutdown complete. Goodbye!
```

In this first example, while the application was able to find the BLE peripheral, it failed to discover the CSC services and characteristics before timing out. Depending on the BLE peripheral, it may take some time before a BLE peripheral advertises both its device services and characteristics. If the peripheral is not responding, you may need to increase the timeout in the `config.toml` file.

```console
 2024/12/16 15:09:47 ----- ----- Starting BLE Sync Cycle 0.8.1
  2024/12/16 15:09:47 [INF] [BLE] created new BLE central controller
  2024/12/16 15:09:47 [INF] [BLE] now scanning the ether for BLE peripheral UUID of F1:42:D8:DE:35:16...
  2024/12/16 15:09:47 [DBG] [BLE] found BLE peripheral F1:42:D8:DE:35:16
  2024/12/16 15:09:47 [DBG] [BLE] connecting to BLE peripheral device F1:42:D8:DE:35:16
  2024/12/16 15:09:47 [INF] [BLE] BLE peripheral device connected
  2024/12/16 15:09:47 [DBG] [BLE] discovering CSC services 00001816-0000-1000-8000-00805f9b34fb
  2024/12/16 15:09:47 [DBG] [BLE] found CSC service 00001816-0000-1000-8000-00805f9b34fb
  2024/12/16 15:09:47 [DBG] [BLE] discovering CSC characteristics 00002a5b-0000-1000-8000-00805f9b34fb
  2024/12/16 15:09:47 [DBG] [BLE] found CSC characteristic 00002a5b-0000-1000-8000-00805f9b34fb
  2024/12/16 15:09:47 [INF] [VID] starting MPV video player...
  2024/12/16 15:09:47 [DBG] [BLE] starting real-time monitoring of BLE sensor notifications...
  2024/12/16 15:09:47 [DBG] [VID] loading video file: cycling_test.mp4
  2024/12/16 15:09:47 [DBG] [VID] entering MPV playback loop...
  2024/12/16 15:09:47 [DBG] [VID] sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
  2024/12/16 15:09:47 [INF] [VID] smoothed sensor speed: 0.00 mph
  2024/12/16 15:09:47 [DBG] [VID] no speed detected, so pausing video
  2024/12/16 15:09:47 [DBG] [VID] video paused successfully
```

In the example above, the application is now running in a loop, periodically querying the BLE peripheral for speed data. The application will also update the video player to match the speed of the sensor. Here, since the video has just begun, its speed is set to 0.0 (paused).

```console
 ...
  2024/12/16 15:13:26 [INF] [SPD] BLE sensor speed: 24.14 mph
  2024/12/16 15:13:26 [DBG] [VID] sensor speed buffer: [0.00 3.17 5.54 13.53 24.14]
  2024/12/16 15:13:26 [INF] [VID] smoothed sensor speed: 9.28 mph
  2024/12/16 15:13:26 [DBG] [VID] last playback speed: 1.74 mph
  2024/12/16 15:13:26 [DBG] [VID] sensor speed delta: 7.54 mph
  2024/12/16 15:13:26 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:26 [INF] [VID] updating video playback speed to 0.56
  2024/12/16 15:13:27 [DBG] [VID] sensor speed buffer: [0.00 3.17 5.54 13.53 24.14]
  2024/12/16 15:13:27 [INF] [VID] smoothed sensor speed: 9.28 mph
  2024/12/16 15:13:27 [DBG] [VID] last playback speed: 9.28 mph
  2024/12/16 15:13:27 [DBG] [VID] sensor speed delta: 0.00 mph
  2024/12/16 15:13:27 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:27 [DBG] [VID] sensor speed buffer: [0.00 3.17 5.54 13.53 24.14]
  2024/12/16 15:13:27 [INF] [VID] smoothed sensor speed: 9.28 mph
  2024/12/16 15:13:27 [DBG] [VID] last playback speed: 9.28 mph
  2024/12/16 15:13:27 [DBG] [VID] sensor speed delta: 0.00 mph
  2024/12/16 15:13:27 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:27 [INF] [SPD] BLE sensor speed: 27.35 mph
  2024/12/16 15:13:27 [DBG] [VID] sensor speed buffer: [3.17 5.54 13.53 24.14 27.35]
  2024/12/16 15:13:27 [INF] [VID] smoothed sensor speed: 14.75 mph
  2024/12/16 15:13:27 [DBG] [VID] last playback speed: 9.28 mph
  2024/12/16 15:13:27 [DBG] [VID] sensor speed delta: 5.47 mph
  2024/12/16 15:13:27 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:27 [INF] [VID] updating video playback speed to 0.88
  ...
```

In this last example, **BLE Sync Cycle** is coordinating with both the BLE peripheral (the speed sensor) and the video player, updating the video player to match the speed of the sensor.

**To quit the application, press `Ctrl+C`.**

```console
 ...
  2024/12/16 15:13:32 [INF] [VID] updating video playback speed to 0.41
  2024/12/16 15:13:32 [DBG] [VID] sensor speed buffer: [11.62 10.58 11.68 0.00 0.00]
  2024/12/16 15:13:32 [INF] [VID] smoothed sensor speed: 6.78 mph
  2024/12/16 15:13:32 [DBG] [VID] last playback speed: 6.78 mph
  2024/12/16 15:13:32 [DBG] [VID] sensor speed delta: 0.00 mph
  2024/12/16 15:13:32 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:33 [INF] [SPD] BLE sensor speed: 12.42 mph
  2024/12/16 15:13:33 [DBG] [VID] sensor speed buffer: [10.58 11.68 0.00 0.00 12.42]
  2024/12/16 15:13:33 [INF] [VID] smoothed sensor speed: 6.94 mph
  2024/12/16 15:13:33 [DBG] [VID] last playback speed: 6.78 mph
  2024/12/16 15:13:33 [DBG] [VID] sensor speed delta: 0.16 mph
  2024/12/16 15:13:33 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:33 [DBG] [VID] sensor speed buffer: [10.58 11.68 0.00 0.00 12.42]
  2024/12/16 15:13:33 [INF] [VID] smoothed sensor speed: 6.94 mph
  2024/12/16 15:13:33 [DBG] [VID] last playback speed: 6.78 mph
  2024/12/16 15:13:33 [DBG] [VID] sensor speed delta: 0.16 mph
  2024/12/16 15:13:33 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:33 [DBG] [VID] sensor speed buffer: [10.58 11.68 0.00 0.00 12.42]
  2024/12/16 15:13:33 [INF] [VID] smoothed sensor speed: 6.94 mph
  2024/12/16 15:13:33 [DBG] [VID] last playback speed: 6.78 mph
  2024/12/16 15:13:33 [DBG] [VID] sensor speed delta: 0.16 mph
  2024/12/16 15:13:33 [DBG] [VID] playback speed update threshold: 0.25 mph
  2024/12/16 15:13:33 [INF] [SPD] BLE sensor speed: 0.00 mph
  2024/12/16 15:13:33 [INF] [VID] user-generated interrupt, stopping video player...
  2024/12/16 15:13:33 [ERR] [APP] context canceled
  2024/12/16 15:13:33 ----- ----- BLE Sync Cycle 0.8.1 shutdown complete. Goodbye!
```

## FAQ

- What is **BLE Sync Cycle**?

> In its simplest form, this application makes video playback run faster when you pedal your bike faster, and slows down video playback when you pedal slower. And, when you stop your bike, video playback pauses.

- How do I use **BLE Sync Cycle**?

> See the [Basic Usage](#basic-usage) section above

- How do I configure **BLE Sync Cycle**?

> See the [Editing the TOML File](#editing-the-toml-file) section above

- Do all Bluetooth devices work with **BLE Sync Cycle**?

> Not necessarily. The Bluetooth package used by **BLE Sync Cycle**, [called Go Bluetooth by TinyGo.org](https://github.com/tinygo-org/bluetooth), is based on the [Bluetooth Low Energy (BLE) standard](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy). Some Bluetooth devices may not be compatible with this protocol.

- Can I disable the log messages in **BLE Sync Cycle**?

> Check out the `logging_level` parameter in the `config.toml` file (see the [Editing the TOML File](#editing-the-toml-file) section above). This parameter can be set to "debug", "info", "warn", or "error", where "debug" is the most verbose (all log messages displayed), and "error" is least verbose.

- My BLE sensor takes a long time to connect, and often times out. What can I do?

> The easiest solution is to just rerun **BLE Sync Cycle**, as that will usually give the BLE sensor enough time to establish a connection. If the issue persists,try increasing the `ble_connect_timeout` parameter in the `config.toml` file (see the [Editing the TOML File](#editing-the-toml-file) section above). Different BLE devices have different advertising intervals, so you may need to adjust this value accordingly.

## Roadmap

Future enhancements and feature requests are now available via the [Github Project page for this repository](https://github.com/users/richbl/projects/4).

## Acknowledgments

- **BLE Sync Cycle** was inspired by indoor cycling training needs, the amazing first-person view cycling videos found on YouTube and elsewhere, and the desire to bring the two together

- A special thanks to the [TinyGo Go Bluetooth package maintainers](https://github.com/tinygo-org/bluetooth) for making BLE device integration in Go relatively straight forward
- Thanks also to the [Gen2Brain team for developing go-mpv](https://github.com/gen2brain/go-mpv), which provides access to the [mpv media player](https://mpv.io/) via their Go package

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
