# BLE Sync Cycle

[![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-ble-sync-cycle)](https://goreportcard.com/report/github.com/richbl/go-ble-sync-cycle)
[![codebeat badge](https://codebeat.co/badges/e840d9ca-937a-41e0-ade2-d2ebf0426720)](https://codebeat.co/projects/github-com-richbl-go-ble-sync-cycle-dev)
![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/richbl/go-ble-sync-cycle?include_prereleases)

## Overview

**BLE Sync Cycle** is a Go application designed to synchronize video playback with real-time cycling data from Bluetooth Low Energy (BLE) devices, such as cycling speed and cadence (CSC) sensors. This integration provides users with an immersive indoor cycling experience by matching video playback with their actual cycling pace, making it a valuable option when outdoor cycling isn't feasible.

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
- Simple command-line interface provides real-time component feedback
- Graceful handling of connection interrupts and system signals ensures all components shut down cleanly

## Rationale

This project was developed to address a specific need: **how can I remain engaged in cycling when the weather outside is less than ideal?**

While there are several existing solutions that allow for "virtual" indoor cycling, such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), these typically require the purchase of specialized training equipment (often preventing the use of your own bike), a subscription to compatible online virtual cycling services, and a reliable broadband Internet connection. My needs, however, are different:

- I want to train using my own bicycle. Since I prefer riding recumbents, it wouldn’t make sense for me to "train" on a traditional upright trainer
- I need a solution that can operate independently without requiring an Internet connection, as I live in a rural area of the Pacific Northwest where both electricity and Internet access can be unreliable at best

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle our regular loss of Internet service

- I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with (I suspect it's my nature as an engineer)

Since I already use a mechanical (no electronics) portable bicycle trainer while riding indoors, it made sense for me to find a way to pair my existing Bluetooth cycling sensors with a local computer which could then drive some kind of interesting feedback while cycling. This project was created to fit that need.

## Requirements

### Hardware Components

- A bicycle set up for indoor riding
- A Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensor, configured for speed
- A computer that supports Bluetooth (4.0+), preferably with a big screen display to watch video playback

For my own indoor cycling configuration, I use a Performance Travel Trac 3 trainer. The BLE sensor used is a [Magene S3+ Speed/Cadence Dual Mode Sensor](https://www.magene.com/en/sensors/59-s3-speed-cadence-dual-mode-sensor.html) configured for speed, though any BLE-compliant sensor should work. For an overview of Bluetooth BLE, refer to the article ["Introduction to Bluetooth Low Energy" by Kevin Townsend](https://learn.adafruit.com/introduction-to-bluetooth-low-energy/introduction). Finally, I'm running **BLE Sync Cycle** on a Lenovo ThinkPad laptop running Ubuntu 24.04 (LTS) connected to a big screen monitor via HDMI.

### Software Components

- The open source, cross-platform [mpv media player](https://mpv.io/), installed and operational
- A local video file for playback using mpv, preferably a first-person view cycling video. Check out [YouTube with this query: "first person cycling"](https://www.youtube.com/results?search_query=first+person+cycling) for some ideas
- This application. While **BLE Sync Cycle** has been written and tested using Ubuntu 24.04 (LTS) on an Intel processor (amd64), it should work across any recent Unix-like platform and architecture
    - In order to compile this project, an operational [Go language](https://go.dev/) environment is required (this release was developed using Go 1.23.2)

## Installation

### Building the Application

1. Clone the repository:

    ```bash
    git clone https://github.com/richbl/go-ble-sync-cycle
    cd go-ble-sync-cycle
    ```

2. Install dependencies:

    ```bash
    go mod download
    ```

3. Build the application:

    ```bash
    go build -o ble-sync-cycle cmd/main.go
    ```

The resulting `build` command will create the`ble-sync-cycle` executable in the current directory.

### Editing the TOML File

Edit the `config.toml` file found in the `internal/configuration` directory. The default file (with a different sensor UUID) is shown below:

```toml
# BLE Sync Cycle TOML configuration
# 0.5.0

[app]
  logging_level = "debug" # Log messages to see during execution: "debug", "info", "warn", "error"
                          # where "debug" is the most verbose and "error" is least verbose

[ble]
  sensor_uuid = "F1:42:D8:DE:35:16" # UUID of BLE peripheral device
  scan_timeout_secs = 30            # Seconds to wait for peripheral response before generating error

[speed]
  smoothing_window = 5          # Number of speed look-backs to use for generating a moving average
  speed_threshold = 1.0         # Minimum speed change to trigger video speed update
  wheel_circumference_mm = 1932 # Wheel circumference in millimeters
  speed_units = "mph"           # "km/h" or "mph"

[video]
  file_path = "cycling_test.mp4" # Path to the video file to play
  display_cycle_speed = true     # Display cycle speed on the screen (true/false)
  display_playback_speed = true  # Display video playback speed on the screen (true/false)
  window_scale_factor = 1.0      # Scale factor for the video window (1.0 = full screen)
  update_interval_sec = 1        # Seconds to wait between video player updates
  speed_multiplier = 0.6         # Multiplier that translates sensor speed (km/h or mph) to video
                                 # playback speed (0.0 = stopped, 1.0 = normal speed)

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
- `display_cycle_speed`: A boolean value that indicates whether to display the cycle sensor speed on the on-screen display (OSD)
- `display_playback_speed`: A boolean value that indicates whether to display the video playback speed on the on-screen display (OSD)
- `window_scale_factor`: A scaling factor for the video window, where 1.0 is full screen. This value can be useful when debugging or when running the video player in a non-maximized window is useful (e.g., 0.5 = half screen)
- `update_interval_sec`: The number of seconds to wait between video player updates
- `speed_multiplier`: The multiplier that translates sensor speed (km/h or mph) to video playback speed (0.0 = stopped, 1.0 = normal speed)

> The `speed_multiplier` parameter is used to control the relative playback speed of the video. Usually, a value of 1.0 is used, as this is the default value (normal playback speed). However, since it's typically unknown what the speed of the bicycle rider in the video is during "normal speed" playback, it's recommended to experiment with different values to find a good balance between  video playback speed and real-world cycling experience.

## Basic Usage

At a high level, **BLE Sync Cycle** will perform the following:

1. Scan for your BLE cycling sensor
2. Connect to the sensor and start receiving speed data
3. Launch MPV video playback
4. Automatically adjust video speed based on your cycling speed
5. Gracefully shutdown on interrupt (Ctrl+C)

To run the application, execute the following command:

```bash
./ble-sync-cycle
```

Or, if the application hasn't yet been built using the `go build` command, you can execute the following command:

```bash
go run cmd/main.go
```

> Be sure that your Bluetooth devices are enabled and in range before running this command. On a computer or similar, you should have your Bluetooth radio turned on. On a BLE sensor, you typically "wake it up" by moving or shaking the device

At this point, you should see the following output:

  ```bash
2024/12/04 16:38:20 Starting BLE Sync Cycle 0.5.0
2024/12/04 16:38:20 INFO [BLE] Created new BLE central controller
2024/12/04 16:38:20 INFO [BLE] Now scanning the ether for BLE peripheral UUID of F1:42:D8:DE:35:16...
2024/12/04 16:38:20 INFO [BLE] Found BLE peripheral F1:42:D8:DE:35:16
2024/12/04 16:38:20 INFO [BLE] Connecting to BLE peripheral device F1:42:D8:DE:35:16
2024/12/04 16:38:24 INFO [BLE] BLE peripheral device connected
2024/12/04 16:38:24 INFO [BLE] Discovering CSC services 00001816-0000-1000-8000-00805f9b34fb
2024/12/04 16:38:34 WARN [BLE] CSC services discovery failed: timeout on DiscoverServices
2024/12/04 16:38:34 ERROR [BLE] BLE peripheral scan failed: timeout on DiscoverServices

  ```

In this first example, while the application was able to find the BLE peripheral, it failed to discover the CSC services and characteristics before timing out. Depending on the BLE peripheral, it may take some time before a BLE peripheral advertises both its device services and characteristics. If the peripheral is not responding, you may need to increase the timeout in the `config.toml` file.

  ```bash
2024/12/04 16:39:07 Starting BLE Sync Cycle 0.5.0
2024/12/04 16:39:07 INFO [BLE] Created new BLE central controller
2024/12/04 16:39:07 INFO [BLE] Now scanning the ether for BLE peripheral UUID of F1:42:D8:DE:35:16...
2024/12/04 16:39:07 INFO [BLE] Found BLE peripheral F1:42:D8:DE:35:16
2024/12/04 16:39:07 INFO [BLE] Connecting to BLE peripheral device F1:42:D8:DE:35:16
2024/12/04 16:39:07 INFO [BLE] BLE peripheral device connected
2024/12/04 16:39:07 INFO [BLE] Discovering CSC services 00001816-0000-1000-8000-00805f9b34fb
2024/12/04 16:39:07 INFO [BLE] Found CSC service 00001816-0000-1000-8000-00805f9b34fb
2024/12/04 16:39:07 INFO [BLE] Discovering CSC characteristics 00002a5b-0000-1000-8000-00805f9b34fb
2024/12/04 16:39:07 INFO [BLE] Found CSC characteristic 00002a5b-0000-1000-8000-00805f9b34fb
2024/12/04 16:39:07 INFO [BLE] Starting real-time monitoring of BLE sensor notifications...
2024/12/04 16:39:07 INFO [VIDEO] Starting MPV video player...
2024/12/04 16:39:07 INFO [VIDEO] Loading video file: cycling_test.mp4
2024/12/04 16:39:07 INFO [VIDEO] Entering MPV playback loop...
2024/12/04 16:39:08 INFO [VIDEO] Sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
2024/12/04 16:39:08 INFO [VIDEO] Smoothed sensor speed: 0.00
2024/12/04 16:39:08 INFO [VIDEO] No speed detected, so pausing video...
2024/12/04 16:39:08 INFO [VIDEO] Video paused successfully
  ```

In the example above, the application is now running in a loop, periodically querying the BLE peripheral for speed data. The application will also update the video player to match the speed of the sensor. Here, since the video has just begun, its speed is set to 0.0 (paused).

```bash
...
2024/12/04 19:31:52 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:31:52 INFO [SPEED] BLE sensor speed: 10.49 mph
2024/12/04 19:31:53 INFO [VIDEO] Sensor speed buffer: [3.17 10.78 10.57 10.34 10.49]
2024/12/04 19:31:53 INFO [VIDEO] Smoothed sensor speed: 9.07
2024/12/04 19:31:53 INFO [VIDEO] Adjusting video speed to 0.54
2024/12/04 19:31:53 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:31:53 INFO [SPEED] BLE sensor speed: 11.10 mph
2024/12/04 19:31:54 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:31:54 INFO [SPEED] BLE sensor speed: 11.91 mph
2024/12/04 19:31:54 INFO [VIDEO] Sensor speed buffer: [10.57 10.34 10.49 11.10 11.91]
2024/12/04 19:31:54 INFO [VIDEO] Smoothed sensor speed: 10.88
2024/12/04 19:31:54 INFO [VIDEO] Adjusting video speed to 0.65
2024/12/04 19:31:54 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:31:54 INFO [SPEED] BLE sensor speed: 12.53 mph
2024/12/04 19:31:55 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:31:55 INFO [SPEED] BLE sensor speed: 12.11 mph
2024/12/04 19:31:55 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:31:55 INFO [SPEED] BLE sensor speed: 12.49 mph
2024/12/04 19:31:55 INFO [VIDEO] Sensor speed buffer: [11.10 11.91 12.53 12.11 12.49]
2024/12/04 19:31:55 INFO [VIDEO] Smoothed sensor speed: 12.02
2024/12/04 19:31:55 INFO [VIDEO] Adjusting video speed to 0.72
...
```

In this last example, **BLE Sync Cycle** is coordinating with both the BLE peripheral (the speed sensor) and the video player, updating the video player to match the speed of the sensor.

**To quit the application, press `Ctrl+C`.**

```bash
...
2024/12/04 19:32:07 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:07 INFO [VIDEO] Sensor speed buffer: [0.00 0.00 8.34 0.00 0.00]
2024/12/04 19:32:07 INFO [VIDEO] Smoothed sensor speed: 1.67
2024/12/04 19:32:07 INFO [VIDEO] Adjusting video speed to 0.10
2024/12/04 19:32:07 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:08 INFO [VIDEO] Sensor speed buffer: [0.00 8.34 0.00 0.00 0.00]
2024/12/04 19:32:08 INFO [VIDEO] Smoothed sensor speed: 1.67
2024/12/04 19:32:08 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:10 INFO [VIDEO] Smoothed sensor speed: 0.00
2024/12/04 19:32:10 INFO [VIDEO] No speed detected, so pausing video...
2024/12/04 19:32:10 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:10 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:11 INFO [VIDEO] Sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
2024/12/04 19:32:11 INFO [VIDEO] Smoothed sensor speed: 0.00
2024/12/04 19:32:11 INFO [VIDEO] No speed detected, so pausing video...
2024/12/04 19:32:11 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:12 INFO [SPEED] Processing speed data from BLE peripheral...
2024/12/04 19:32:12 INFO [VIDEO] Sensor speed buffer: [0.00 0.00 0.00 0.00 0.00]
2024/12/04 19:32:12 INFO [VIDEO] Smoothed sensor speed: 0.00
2024/12/04 19:32:12 INFO [VIDEO] No speed detected, so pausing video...
2024/12/04 19:32:12 INFO [SPEED] Processing speed data from BLE peripheral...
^C2024/12/04 19:32:12 INFO [APP] Shutdown signal received
2024/12/04 19:32:12 INFO [VIDEO] Context cancelled. Shutting down video player component
2024/12/04 19:32:12 INFO [APP] Application shutdown complete. Goodbye!
```

## FAQ

Q: What is **BLE Sync Cycle**?
A: In its simplest form, this application makes video playback run faster when you pedal your bike faster, and slows down video playback when you pedal slower. And, when you stop your bike, video playback pauses.

Q: Do all Bluetooth devices work with **BLE Sync Cycle**?
A: Not necessarily. The Bluetooth package used by **BLE Sync Cycle**, [called Go Bluetooth by TinyGo.org](https://github.com/tinygo-org/bluetooth), is based on the [Bluetooth Low Energy (BLE) standard](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy). Some Bluetooth devices may not be compatible with this protocol.

Q: Can I disable the log messages in **BLE Sync Cycle**?
A: While you cannot entirely disable log messages, check out the `logging_level` parameter in the `config.toml` file (see the [Editing the TOML File](#editing-the-toml-file) section above). This parameter can be set to "debug", "info", "warn", or "error", where "debug" is the most verbose and "error" is least verbose. When set to "error", only error/fatal messages will be displayed which, under normal circumstances, should be none.

Q: How do I use **BLE Sync Cycle**?
A: See the [Basic Usage](#basic-usage) section above

Q: How do I configure **BLE Sync Cycle**?
A: See the [Editing the TOML File](#editing-the-toml-file) section above

## Roadmap

Future enhancements include (in no particular order):

- Add support for other video players (e.g., VLC)
- Add optional check for battery status of BLE peripheral device
- Create a desktop application (GUI) for **BLE Sync Cycle**
- Add support for non-BLE peripheral devices
- Automatically quit the application after a period of inactivity
- As an exercise, refactor using [the Rust language](https://www.rust-lang.org/)

## Acknowledgments

- **BLE Sync Cycle** was inspired by indoor cycling training needs, the amazing first-person view cycling videos found on YouTube and elsewhere, and the desire to bring the two together

- A special thanks to the [TinyGo Go Bluetooth package maintainers](https://github.com/tinygo-org/bluetooth) for making BLE device integration in Go relatively straight forward
- Thanks also to the [Gen2Brain team for developing go-mpv](https://github.com/gen2brain/go-mpv), which provides access to the [mpv media player](https://mpv.io/) via their Go package

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
