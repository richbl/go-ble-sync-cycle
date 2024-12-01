# BLE Sync Cycle

[![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-ble-sync-cycle)](https://goreportcard.com/report/github.com/richbl/go-ble-sync-cycle)
[![codebeat badge](https://codebeat.co/badges/155e9293-7023-4956-81f5-b3cde7b93842)](https://codebeat.co/projects/github-com-richbl-go-ble-sync-cycle-main)
![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/richbl/go-ble-sync-cycle?include_prereleases)

## Overview

**BLE Sync Cycle** is a Go application that synchronizes video playback with wireless real-time cycling data obtained from a Bluetooth Low Energy (BLE) device such as a cycling speed and cadence (CSC) sensor. This allows users to create an immersive indoor cycling experience by matching video playback to their actual cycling pace, providing an engaging experience when cycling outdoors isn't an option.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><img src="[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9)" width=700></picture>
</p>

## 🚲 Features

- Real-time synchronization between cycling speed and video playback
- Support for compliant Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensors
- TOML-based configuration for easy application customization that includes:
    - BLE sensor identification (UUID)
    - Bluetooth device scanning timeout
    - Wheel circumference for accurate speed conversion
    - Support for different speed units: miles per hour (mph), kilometers per hour (kph)
    - Speed smoothing for natural and seamless video playback
    - Choice of video file for playback
- Simple command-line interface provides real-time component feedback
- Graceful handling of connection interrupts and system signals ensures all components shut down cleanly

## Rationale

This project was created to serve a very specific need: **how can I continue to stay engaged in cycling when the weather outside is not particularly cooperative?**

While there are quite a few existing solutions that permit me to cycle indoors "virtually," such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), they typically require the purchase of dedicated training hardware (often preventing you from using your own bike), a compatible online virtual cycling services membership subscription, and a stable broadband Internet connection. My requirements were different:

- I want to use my own bicycle while training: since I prefer riding recumbents, it'd make zero sense for me to "train" on a purpose-built upright trainer
- I want a solution that can run standalone and doesn't require an Internet connection, as I live in a rural part of the Pacific Northwest where both electricity and Internet access are temperamental at best

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle our regular loss of Internet service

- I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with (I suspect it's my nature as an engineer)

Since I already use an old (and very analog) portable fluid bicycle trainer while riding indoors, it made sense to find a way to pair my existing Bluetooth cycling sensors with a local computer which could then drive some kind of interesting feedback while cycling. This project was created to fit that need.

## 📋 Requirements

### Hardware Components

- A bicycle set up for indoor riding
- A Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensor, configured for speed
- A computer that supports Bluetooth (4.0+), preferably with a big screen display to watch video playback

For my own indoor cycling configuration, I use an old _Performance Travel Trac 3_ fluid trainer. The BLE sensor used is a [Magene S3+ Speed/Cadence Dual Mode Sensor](https://www.magene.com/en/sensors/59-s3-speed-cadence-dual-mode-sensor.html) configured for speed, though any BLE-compliant sensor should work. For an overview of Bluetooth BLE, [read the article "Introduction to Bluetooth Low Energy" by Kevin Townsend](https://learn.adafruit.com/introduction-to-bluetooth-low-energy/introduction). Finally, I'm running **BLE Sync Cycle** on a Lenovo ThinkPad T15 laptop running Ubuntu 24.04 (LTS) connected to a big screen monitor via HDMI.

### Software Components

- The open source, cross-platform [mpv media player](https://mpv.io/), installed and operational
- A local video file for playback using mpv, preferably a first-person view cycling video. Check out [YouTube with the query "first person cycling"](https://www.youtube.com/results?search_query=first+person+cycling) for some ideas
- This application. While **BLE Sync Cycle** has been written and tested using Ubuntu 24.04 (LTS) on an Intel Core i7 processor, it should work on any recent Unix-like platform
    - In order to compile this project, an operational [Go language](https://go.dev/) environment is required (this release was developed using Go 1.23.2)

## 🛠️ Installation

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
# TOML configuration file for the ble-sync-cycle application

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
  update_interval_sec = 1         # Seconds to wait between video player updates
  speed_multiplier = 0.6          # Multiplier that translates sensor speed (km/h or mph) to
                                  # video playback speed (0.0 = stopped, 1.0 = normal speed)

```

An explanation of the various sections of the `config.toml` file is provided below:

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
- `update_interval_sec`: The number of seconds to wait between video player updates
- `speed_multiplier`: The multiplier that translates sensor speed (km/h or mph) to video playback speed (0.0 = stopped, 1.0 = normal speed)

> The `speed_multiplier` parameter is used to control the relative playback speed of the video. Usually, a value of 1.0 is used, as this is the default value (normal playback speed). However, since it's typically unknown what the speed of the bicycle rider in the video is during "normal speed" playback, it's recommended to experiment with different values to find a good balance between  video playback speed and real-world cycling experience.

## Basic Usage

At a high level, **BLE Sync Cycle** will do the following:

1. Scan for your BLE cycling sensor
2. Connect to the sensor and start receiving speed data
3. Launch video playback
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

> Be sure that your Bluetooth devices are enabled before running this command. On a computer or similar, you should have your Bluetooth radio turned on. On a BLE sensor, you typically "wake it up" by moving or shaking the device.

At this point, you should see the following output:

  ```bash
2024/11/30 17:28:50 \ Created new BLE central controller
2024/11/30 17:28:50 \ Now scanning the ether for BLE peripheral UUID of F1:42:D8:DE:35:16 ...
2024/11/30 17:28:50 \ Found BLE peripheral F1:42:D8:DE:35:16
2024/11/30 17:28:50 \ Connecting to BLE peripheral device F1:42:D8:DE:35:16
2024/11/30 17:28:53 \ BLE peripheral device connected
2024/11/30 17:28:53 \ Discovering CSC services 00001816-0000-1000-8000-00805f9b34fb
2024/11/30 17:29:03 \ CSC services discovery failed: timeout on DiscoverServices
2024/11/30 17:29:03 \ BLE peripheral scan failed: timeout on DiscoverServices
  ```

In this first instance, while the application was able to find the BLE peripheral, it failed to discover the CSC services and characteristics before timing out. Depending on the BLE peripheral, it may take some time before a BLE peripheral advertises both its device services and characteristics. If the peripheral is not responding, you may need to increase the timeout in the `config.toml` file.

  ```bash
2024/11/30 17:30:37 \ Created new BLE central controller
2024/11/30 17:30:37 \ Now scanning the ether for BLE peripheral UUID of F1:42:D8:DE:35:16 ...
2024/11/30 17:30:37 \ Found BLE peripheral F1:42:D8:DE:35:16
2024/11/30 17:30:37 \ Connecting to BLE peripheral device F1:42:D8:DE:35:16
2024/11/30 17:30:37 \ BLE peripheral device connected
2024/11/30 17:30:37 \ Discovering CSC services 00001816-0000-1000-8000-00805f9b34fb
2024/11/30 17:30:37 \ Found CSC service 00001816-0000-1000-8000-00805f9b34fb
2024/11/30 17:30:37 \ Discovering CSC characteristics 00002a5b-0000-1000-8000-00805f9b34fb
2024/11/30 17:30:37 \ Found CSC characteristic 00002a5b-0000-1000-8000-00805f9b34fb
2024/11/30 17:30:37 \ Starting real-time monitoring of BLE sensor notifications...
2024/11/30 17:30:37 / Starting MPV video player...
2024/11/30 17:30:37 / Loading video file: cycling_test.mp4
2024/11/30 17:30:37 / Entering MPV playback loop...
2024/11/30 17:30:38 / Current sensor speed: 0.00 ... Last sensor speed: 0.00
2024/11/30 17:30:38 / No speed detected, so pausing video...
2024/11/30 17:30:38 / Video paused successfully
  ```

In the example above, the application is now running in a loop, periodically querying the BLE peripheral for speed data. The application will also update the video player to match the speed of the sensor. Here, since the video has just begun, its speed is set to 0.0 (paused).

```bash
  ...
2024/11/30 17:32:03 | Processing speed data from BLE peripheral...
2024/11/30 17:32:03 | BLE sensor speed: 16.19 mph
2024/11/30 17:32:03 / Current sensor speed: 17.46 ... Last sensor speed: 16.97
2024/11/30 17:32:04 | Processing speed data from BLE peripheral...
2024/11/30 17:32:04 | BLE sensor speed: 14.7 mph
2024/11/30 17:32:04 / Current sensor speed: 17.01 ... Last sensor speed: 17.46
2024/11/30 17:32:05 | Processing speed data from BLE peripheral...
2024/11/30 17:32:05 | BLE sensor speed: 14.38 mph
2024/11/30 17:32:05 | Processing speed data from BLE peripheral...
2024/11/30 17:32:05 | Processing speed data from BLE peripheral...
2024/11/30 17:32:05 / Current sensor speed: 9.05 ... Last sensor speed: 17.01
2024/11/30 17:32:05 / Adjusting video speed to 0.54
2024/11/30 17:32:05 / Video speed updated successfully
2024/11/30 17:32:06 | Processing speed data from BLE peripheral...
2024/11/30 17:32:06 | BLE sensor speed: 13.9 mph
2024/11/30 17:32:06 / Current sensor speed: 8.60 ... Last sensor speed: 9.05
2024/11/30 17:32:06 | Processing speed data from BLE peripheral...
2024/11/30 17:32:07 | Processing speed data from BLE peripheral...
2024/11/30 17:32:07 / Current sensor speed: 2.78 ... Last sensor speed: 8.60
2024/11/30 17:32:07 / Adjusting video speed to 0.17
2024/11/30 17:32:07 / Video speed updated successfully
...

```

In this last example, **BLE Sync Cycle** is coordinating with both the BLE peripheral (the speed sensor) and the video player, updating the video player to match the speed of the sensor.

**To quit the application, press `Ctrl+C`.**

```bash
...
2024/11/30 17:32:06 | Processing speed data from BLE peripheral...
2024/11/30 17:32:06 | BLE sensor speed: 13.9 mph
2024/11/30 17:32:06 / Current sensor speed: 8.60 ... Last sensor speed: 9.05
2024/11/30 17:32:06 | Processing speed data from BLE peripheral...
2024/11/30 17:32:07 | Processing speed data from BLE peripheral...
2024/11/30 17:32:07 / Current sensor speed: 2.78 ... Last sensor speed: 8.60
2024/11/30 17:32:07 / Adjusting video speed to 0.17
2024/11/30 17:32:07 / Video speed updated successfully
2024/11/30 17:32:07 | Processing speed data from BLE peripheral...
2024/11/30 17:32:07 | BLE sensor speed: 6.03 mph
2024/11/30 17:32:07 | Processing speed data from BLE peripheral...
^C2024/11/30 17:32:07 / Context cancelled. Shutting down video player component
2024/11/30 17:32:07 - Shutdown signal received
2024/11/30 17:32:07 - Application shutdown complete. Goodbye!
```

## ⚠️ FAQ

Q: What is **BLE Sync Cycle**?
A: In its simplest form, this application makes video playback run faster when you pedal your bike faster, and slows down video playback when you pedal slower. And, when you stop your bike, video playback pauses.

Q: Do all Bluetooth devices work with **BLE Sync Cycle**?
A: Not necessarily. The Bluetooth package used by **BLE Sync Cycle**, [called Go Bluetooth by TinyGo.org](https://github.com/tinygo-org/bluetooth), is based on the [Bluetooth Low Energy (BLE) standard](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy). Some Bluetooth devices may not be compatible with this protocol.

Q: Can I disable the log messages in **BLE Sync Cycle**?
A: Yes, but it's currently not a setting that can be changed in the `config.toml` file. Instead, you'll need to modify the source code directly (check out the `//Disable logging` comment in `main.go`).

Q: How do I use **BLE Sync Cycle**?
A: See the [Basic Usage](#basic-usage) section above

Q: How do I configure **BLE Sync Cycle**?
A: See the [Editing the TOML File](#editing-the-toml-file) section above

## Roadmap

- Add support for other video players (e.g., VLC)
- Add flag in `config.toml` to enable/disable logging messages
- Add optional check for battery status of BLE peripheral device
- Add support for non-BLE peripheral devices

## 🙏 Acknowledgments

- Thanks to the [TinyGo Go Bluetooth package maintainers](https://github.com/tinygo-org/bluetooth)
- Inspired by indoor cycling training needs, the amazing first-person view cycling videos found on YouTube and elsewhere, and the desire to bring the two together

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
