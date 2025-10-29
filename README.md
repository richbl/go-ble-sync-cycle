<p align="left">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=350></picture>
</p>

![GitHub Release](https://img.shields.io/github/v/release/richbl/go-ble-sync-cycle?include_prereleases&sort=semver&display_name=tag&color=blue) [![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-ble-sync-cycle)](https://goreportcard.com/report/github.com/richbl/go-ble-sync-cycle) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/595889e53f25475da18dea64b5a60419)](https://app.codacy.com/gh/richbl/go-ble-sync-cycle/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=richbl_go-ble-sync-cycle&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=richbl_go-ble-sync-cycle)

## Overview

**BLE Sync Cycle** is a Go application designed to synchronize video playback with real-time cycling data from Bluetooth Low Energy (BLE) devices, such as cycling speed and cadence (CSC) sensors. This integration provides users with a more immersive indoor cycling experience by matching video playback speed with their actual cycling pace, making it a great option when outdoor cycling isn't feasible.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9"><img src="[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9)" width=700></picture>
</p>

## Features

* Real-time synchronization of cycling speed and video playback

* Supports compliant BLE Cycling Speed and Cadence (CSC) sensors (in speed mode)

* Integrates with [mpv](https://mpv.io) and [VLC](https://www.videolan.org) media players

* Highly configurable TOML-based config file for:
    * BLE sensor address (BD\_ADDR) and scan timeout
    * Wheel circumference (for accurate speed)
    * Speed units (mph or km/h)
    * Speed smoothing for natural playback
    * Video file selection
    * Display options:
        * On-screen display (OSD) for speed and time remaining
        * Video window scaling (fullscreen, etc.)
        * OSD position and font size

* Command-line interface for real-time application status

* CLI flags to override settings:
    * Configuration file path (allows for multiple profiles)
    * Video start time (seek)
    * Help/usage information

* Configurable log levels (debug, info, warn, error)

* On every application startup, the battery level of the BLE sensor is checked and displayed

* Graceful handling of connection interrupts and system signals for a clean shutdown

## Rationale

This project was developed to address a specific need: **how can I continue cycling when the weather outside is less than ideal?**

While there are several existing solutions that allow for "virtual" indoor cycling, such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), these typically require the purchase of specialized training equipment (often preventing the use of your own bike), a subscription to compatible online virtual cycling services, and a reliable broadband Internet connection.

My needs are different:

* I want to train _using my own bicycle_. Since I prefer riding recumbents, it wouldn’t make sense for me to train on a traditional upright trainer
* I need a solution that can function with minimal dependencies and without requiring an Internet connection, as I live in a rural part of the Pacific Northwest where both electrical and Internet services are unreliable

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle our regular loss of Internet service here in the woods of the Pacific Northwest

* Finally (and importantly), I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with. Call me crazy, but I suspect it's my nature as an engineer to tinker...

Since I already use an analog bicycle trainer while riding indoors, it made sense for me to find a way to pair my existing Bluetooth cycling sensors with a local computer which could then drive some kind of interesting feedback while cycling. This project was created to fit that need.

<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6"><img src="[[https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6)]([https://github.com/user-attachments/assets/a3165440-33d8-42a9-9992-8acf18375da9](https://github.com/user-attachments/assets/b33d68ac-0e4e-42b0-8d08-d4d5dac0cde6))" width=700></picture>
</p>

## Would You Like to Know More?

For more information about **BLE Sync Cycle**, check out the [BLE Sync Cycle project wiki](https://github.com/richbl/go-ble-sync-cycle/wiki). The wiki includes the following sections:

* BLE Sync Cycle
    * Home
    * Features
    * Rationale

* Requirements
    * Hardware
    * Software

* Installation
    * Application Dependencies
    * Building the Application
    * Editing the TOML File
        * The [app] Section
        * The [ble] Section
        * The [speed] Section
        * The [video] Section
        * The [video.OSD] Section

* Basic Usage
    * Overview
    * Running the Application
    * Using the Command Line Options
        * Setting the Configuration File Path
        * Seeking to a Specific Time in the Video
        * Displaying Help in BLE Sync Cycle

* FAQ
    * Frequently Asked Questions

* Roadmap
    * Roadmap

* Acknowledgements
    * Acknowledgements

* License
    * License

Enjoy!
