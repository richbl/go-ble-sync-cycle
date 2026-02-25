<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

* Real-time synchronization of cycling speed and video playback

* Support for compliant BLE Cycling Speed and Cadence (CSC) sensors (in speed mode)

* Integrates with the [mpv](https://mpv.io) media player

* Highly configurable TOML-based configuration files for:
    * BLE sensor address (BD\_ADDR) and scan timeout
    * Wheel circumference (for accurate speed)
    * Speed units (mph or km/h)
    * Speed smoothing for natural playback
    * Video file selection with support for multiple file formats (mp4, mkv, etc.)
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

* The battery level of the BLE sensor is checked and displayed on every session start, ensuring that users have sufficient battery life for their training session

* Graceful handling of connection interrupts and system signals for a clean shutdown

<!-- markdownlint-disable MD033 -->
<p align="center">
<img width="850" alt="Screenshot showing BSC GUI" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/gui_2x2_part1.png">
</p>

<p align="center">
<img width="850" alt="Screenshot showing BSC GUI" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/gui_2x2_part2.png">
</p>
<!-- markdownlint-enable MD033 -->
