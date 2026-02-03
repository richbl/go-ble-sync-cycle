<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="left">
<img width="350" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

![GitHub Release](https://img.shields.io/github/v/release/richbl/go-ble-sync-cycle?include_prereleases&sort=semver&display_name=tag&color=blue) [![Go Report Card](https://goreportcard.com/badge/github.com/richbl/go-ble-sync-cycle)](https://goreportcard.com/report/github.com/richbl/go-ble-sync-cycle) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/595889e53f25475da18dea64b5a60419)](https://app.codacy.com/gh/richbl/go-ble-sync-cycle/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=richbl_go-ble-sync-cycle&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=richbl_go-ble-sync-cycle)

## Overview

**BLE Sync Cycle** is a Go application designed to synchronize video playback with real-time cycling data from Bluetooth Low Energy (BLE) devices, such as cycling speed and cadence (CSC) sensors. This integration provides users with a more immersive and engaging indoor cycling experience by matching first-person video playback speed with their actual cycling pace.

<!-- markdownlint-disable MD033 -->
<p align="center">
<img width="850" alt="Screenshot showing cycling trainer" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/trainer_0_hd.png">
</p>

<p align="center">
<img width="850" alt="Screenshot showing BSC GUI" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/trail_gui_hd.png">
</p>

Here's a short (~30 seconds) YouTube video demonstrating how BLE Sync Cycle works:

<p align="center">
  <a href="https://youtu.be/oZqs__8KdnI"><img width="850" alt="Screenshot showing YouTube thumbnail" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/bsc_video_thumbnail.png"></a>
</p>
<!-- markdownlint-enable MD033 -->

## Features

* Real-time synchronization of cycling speed and video playback

* Support for compliant BLE Cycling Speed and Cadence (CSC) sensors (in speed mode)

* Integrates with [mpv](https://mpv.io) and [VLC](https://www.videolan.org) media players

* Highly configurable TOML-based configuration files for:
    * BLE sensor address (BD\_ADDR) and scan timeout
    * Wheel circumference (for accurate speed)
    * Speed units (mph or km/h)
    * Speed smoothing for natural playback
    * Choice of media player (mpv or VLC)
    * Video file selection
    * Seek to a specific start time in the video
    * Display options:
        * On-screen display (OSD) for speed and time remaining
        * Video window scaling (fullscreen, etc.)
        * OSD position and font size

* Choice of running modes:
    * GUI Mode: a modern GTK4/Adwaita design with interactive graphical support for:
        * Cycling session selection
        * Session status (including cycling speed and session time remaining), and video playback
        * Session logging
        * Session editing and management

    * CLI Mode: a simple command-line interface for real-time application status with minimal operational overhead
        * Application flags to override configuration file settings:
            * Configuration file path (allows for multiple profiles)
            * Video start time (seek)
            * Help/usage information

* Configurable logging levels (debug, info, warn, error)

* On every application startup, the battery level of the BLE sensor is checked and displayed

* Graceful handling of connection interrupts and system signals for a clean shutdown

<!-- markdownlint-disable MD033 -->
<p align="center">
<img width="850" alt="Screenshot showing BSC GUI" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/gui_2x2_part1.png">
</p>

<p align="center">
<img width="850" alt="Screenshot showing BSC GUI" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/gui_2x2_part2.png">
</p>
<!-- markdownlint-enable MD033 -->

## Rationale

This project was developed to address a specific need:

**How can I continue cycling when the weather outside is not ideal?**

While there are several existing solutions that allow for virtual indoor cycling, such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), these typically require the purchase of specialized training equipment (often precluding the use of your own bike), a subscription to compatible online virtual cycling services, and a reliable broadband Internet connection.

My needs are different:

* I want to train _using my own bicycle_. In my own case, I prefer riding recumbents, so it wouldn’t make sense for me to train on a traditional upright trainer
* I need a solution that can operate with minimal dependencies and without requiring an Internet connection, as I live in a rural part of the Pacific Northwest where both electricity and Internet services can be unreliable

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle the notification of our regular loss of Internet service here in the woods of the Pacific Northwest

* Finally, I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with. I suspect it's my nature as an engineer to tinker...

Since I already use an analog bicycle trainer while riding indoors, it made sense for me to find a way to pair Bluetooth cycling sensors with a local computer which could then drive some kind of interesting video feedback while cycling. This project was created to fit that need.

<!-- markdownlint-disable MD033 -->
<p align="center">
<img width="850" alt="Screenshot showing cycling trainer" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/trainer_1_hd.png">
</p>
<!-- markdownlint-enable MD033 -->

## Would You Like to Know More?

For more information about **BLE Sync Cycle**, check out the [BLE Sync Cycle project wiki](https://github.com/richbl/go-ble-sync-cycle/wiki). The wiki includes the following sections:

### BLE Sync Cycle

* Home
* Features
* Rationale

#### Requirements

* Hardware
* Software

#### Installation

* Application Dependencies
* Building the Application
* Installing the Application

#### Basic Usage

* Overview
* Running the Application
    * GUI Mode
    * CLI Mode
        * Using the Command Line Options
            * Setting the Configuration File Path
            * Seeking to a Specific Time in the Video
            * Displaying Help in BLE Sync Cycle
* Anatomy of a BSC TOML File
    * The App Section
    * The BLE Section
    * The Speed Section
    * The Video Section
    * The Video On-Screen Display Section
* FAQ
* Roadmap
* Acknowledgements
* License

Enjoy!
