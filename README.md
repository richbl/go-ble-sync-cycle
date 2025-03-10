# BLE Sync Cycle

![GitHub Release](https://img.shields.io/github/v/release/richbl/go-ble-sync-cycle?include_prereleases&sort=semver&display_name=tag&color=blue) [![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-ble-sync-cycle)](https://goreportcard.com/report/github.com/richbl/go-ble-sync-cycle) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/595889e53f25475da18dea64b5a60419)](https://app.codacy.com/gh/richbl/go-ble-sync-cycle/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=richbl_go-ble-sync-cycle&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=richbl_go-ble-sync-cycle)

## Overview

**BLE Sync Cycle** is a Go application designed to synchronize video playback with real-time cycling data from Bluetooth Low Energy (BLE) devices, such as cycling speed and cadence (CSC) sensors. This integration provides users with a more immersive indoor cycling experience by matching video playback with their actual cycling pace, making it a great option when outdoor cycling isn't feasible.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><img src="[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9)" width=700></picture>
</p>

## Features

- Real-time synchronization between cycling speed and video playback
- Support for compliant Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensors (configured for speed mode)
- TOML-based configuration for application customizations that include:
    - BLE sensor setup (BD_ADDR)
    - Bluetooth device scanning timeout
    - Wheel circumference, required for accurate speed conversion
    - Support for different speed units: miles per hour (mph) and kilometers per hour (km/h)
    - Speed smoothing option for a more natural video playback
    - Configurable choice of video file for playback
    - Various display options for optimal video playback, including:
        - The display of sensor speed, video playback speed and playback time remaining via on-screen display (OSD)
        - Video window scaling (full screen, half screen, etc.)
        - OSD font size
- Simple command-line interface provides real-time application status
    - Command-line flag options provide for easy override of configuration settings, including:
        - Location of the configuration file
            - Allowing for the creation of multiple configuration files that can be created to support different cycling sessions and different bicycle configurations
        - Where to start video playback (seek functionality)
        - Display of application usage/help information
- Configurable logging levels (debug, info, warn, error) to manage the information displayed during application execution
- Graceful handling of connection interrupts and system signals ensures all components shut down cleanly upon application exit

## Rationale

This project was developed to address a specific need: **how can I continue cycling when the weather outside is less than ideal?**

While there are several existing solutions that allow for "virtual" indoor cycling, such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), these typically require the purchase of specialized training equipment (often preventing the use of your own bike), a subscription to compatible online virtual cycling services, and a reliable broadband Internet connection.

My needs are a bit different:

- I want to train _using my own bicycle_. Since I prefer riding recumbents, it wouldn’t make sense for me to train on a traditional upright trainer
- I need a solution that can function with minimal dependencies and without requiring an Internet connection, as I live in a rural part of the Pacific Northwest where both electricity and Internet access can be unreliable at best

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle our regular loss of Internet service here in the woods of the Pacific Northwest

- I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with. Call me crazy, but I suspect it's my nature as an engineer to tinker...

Since I already use a (non-digital) bicycle trainer while riding indoors, it made sense for me to find a way to pair my existing Bluetooth cycling sensors with a local computer which could then drive some kind of interesting feedback while cycling. This project was created to fit that need.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6"><img src="[[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6)]([https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6))" width=700></picture>
</p>

## Want to Know More?

For more information about **BLE Sync Cycle**, check out the [BLE Sync Cycle project wiki](https://github.com/richbl/go-ble-sync-cycle/wiki). The wiki includes the following sections:

- Hardware and software requirements
- Application installation
    - Configuring the application to best suit your own needs
- Running the application
    - Stepping through an example application startup)
- Frequently Asked Questions (FAQ)
- Project roadmap
- Acknowledgements
    - A big thanks to the various package owners that made this project possible
- Project license

Enjoy!