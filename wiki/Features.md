<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

- Real-time synchronization between cycling speed and video playback
- Support for compliant Bluetooth Low Energy (BLE) Cycling Speed and Cadence (CSC) sensors (configured for speed mode)
- Support for multiple media players, including [mpv](https://mpv.io) and [VLC](https://www.videolan.org)
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
        - Position of the on-screen OSD
        - OSD font size
- Simple command-line interface provides real-time application status
    - Command-line flag options provide for easy override of configuration settings, including:
        - Location of the configuration file
            - Allowing for the creation of multiple configuration files that can be created to support different cycling sessions and different bicycle configurations
        - Where to start video playback (seek functionality)
        - Display of application usage/help information
- Configurable logger levels (debug, info, warn, error) to manage the information displayed during application execution
- Graceful handling of connection interrupts and system signals ensures all components shut down cleanly upon application exit
