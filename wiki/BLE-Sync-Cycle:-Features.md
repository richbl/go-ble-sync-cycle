<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

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
    * GUI Mode: modern GTK4/Adwaita design with full support for in-application session selection, playback, logging and session editing
    * CLI Mode: a simple command-line interface for real-time application status with minimal operational overhead

        * Application flags to override configuration file settings:
            * Configuration file path (allows for multiple profiles)
            * Video start time (seek)
            * Help/usage information

* Configurable logging levels (debug, info, warn, error)

* On every application startup, the battery level of the BLE sensor is checked and displayed

* Graceful handling of connection interrupts and system signals for a clean shutdown
