<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

Edit the `config.toml` file found in the `internal/config` directory. With the exception of updating `sensor_bd_addr` to the actual sensor ID to be used, these default settings should be appropriate for most application use cases.

The default `config.toml` file is shown below:

```toml
# BLE Sync Cycle TOML configuration
# 0.12.0

[app]
  logging_level = "info" # Log messages to see during execution: "debug", "info", "warn", "error"
                          # where "debug" is the most verbose and "error" is least verbose

[ble]
  sensor_bd_addr = "F1:42:D8:DE:35:16" # Address of BLE peripheral device (e.g. "11:22:33:44:55:66")
  scan_timeout_secs = 30               # Seconds (1-100) to wait for peripheral response before generating error

[speed]
  speed_threshold = 0.25        # Minimum speed change (0.00-10.00) to trigger video speed update
  speed_units = "mph"           # "km/h" or "mph"
  smoothing_window = 5          # Number of recent speeds (1-25) to generate a moving average
  wheel_circumference_mm = 2155 # Wheel circumference in millimeters (50-3000) 

[video]
  media_player = "mpv"           # Media player to use ("mpv" or "vlc")
  file_path = "cycling_test.mp4" # Path to the video file to play
  window_scale_factor = 1.0      # Scale factor (0.1-1.0) for the video window (1.0 = full screen)
  seek_to_position = "00:00"     # Seek minutes:seconds ("MM:SS") into the video playback
  update_interval_sec = 0.25     # Seconds (0.10-3.00) to wait between video player updates
  speed_multiplier = 0.8         # Multiplier (0.1-1.5) that adjusts sensor speed to video playback speed
                                 # (0.1 = slower, 1.0 = normal, 1.5 = faster playback)
  [video.OSD]
    display_cycle_speed = true    # Display cycle speed on the on-screen display (true/false)
    display_playback_speed = true # Display video playback speed on the on-screen display (true/false)
    display_time_remaining = true # Display time remaining on the on-screen display (true/false)
    font_size = 40                # Font size (10-200) for on-screen display (OSD)
    margin_left = 25              # pixel offset of OSD (0-100) from the left of the media player window
    margin_top = 25               # pixel offset of the OSD (0-100) from the top of the media player window
```

An explanation of the various sections of the `config.toml` file is provided below:

## The `[app]` Section

The `[app]` section is used for configuration of the **BLE Sync Cycle** application itself. It includes the following parameter:

- `logging_level`: The logging level to use, which displays messages to the console as the application executes. This can be "debug", "info", "warn", or "error", where "debug" is the most verbose and "error" is least verbose.

## The `[ble]` Section

The `[ble]` section configures your computer (referred to as the BLE central controller) to scan for and query the BLE speed sensor (referred to as the BLE peripheral). It includes the following parameters:

- `sensor_bd_addr`: The address of the BLE peripheral device (e.g., sensor) to connect with and monitor for speed data
- `scan_timeout_secs`: The number of seconds to wait for a BLE peripheral response before generating an error message. Some BLE devices can take a while to respond (called "advertising"), so adjust this value accordingly. A value of 30 seconds is a good starting point.

> To find the address (BD_ADDR) of your BLE peripheral device, you'll need to connect to it from your computer (or any device with Bluetooth connectivity). From Ubuntu (or any other Linux distribution), you can use [the `bluetoothctl` command](https://www.mankier.com/1/bluetoothctl#). BLE peripheral device BD_ADDRs are typically in the form of "11:22:33:44:55:66."

## The `[speed]` Section

The `[speed]` section defines the configuration for the speed controller component. The speed controller takes raw BLE CSC speed data (a rate of discrete device events per time cycle) and converts it speed (either km/h or mph, depending on `speed_units`). It includes the following parameters:

- `speed_threshold`: The minimum speed change to trigger video speed updates
- `speed_units`: The speed units to use (either "km/h" or "mph")
- `smoothing_window`: The number of "look-backs" (most recent speed measurements) to use for generating a moving average for the speed value
- `wheel_circumference_mm`: The wheel circumference in millimeters: important in order to accurately convert raw sensor values to actual speed (distance traveled per unit time)

> The smoothing window is a simple ring buffer that stores the last (n) speed measurements, meaning that it will create a moving average for the speed value. This helps to smooth out the speed data and provide a more natural video playback experience.

## The `[video]` Section

The `[video]` section defines the configuration for the MPV video player component. It includes the following parameters:

- `media_player`: The media player to use (only "mpv" is currently supported)
- `file_path`: The full path to the video file to play. The video format must be supported by MPV (e.g., MP4, webm, etc.)
- `window_scale_factor`: A scaling factor for the video window, where 1.0 is full screen. This value can be useful when debugging or when running the video player in a non-maximized window is preferred
- `seek_to_position`: The minutes:seconds ("MM:SS") to seek to a specific point in video playback. This can be useful with longer videos that may take multiple training sessions to complete
- `update_interval_sec`: The number of seconds to wait between video player updates

> The `speed_multiplier` parameter is used to control the relative playback speed of the video. Usually, a value of 1.0 is used (<1.0 will slow playback; >1.0 will speed up playback), as this is the default value (normal playback speed). However, since it's typically unknown what the speed of the vehicle is in the video during "normal speed" playback, it's recommended to experiment with different values to find a good balance between video playback speed and real-world cycling experience.

## The `[video.OSD]` Section

- `display_cycle_speed`: A boolean value that indicates whether to display the cycle sensor speed on the on-screen display (OSD)
- `display_playback_speed`: A boolean value that indicates whether to display the video playback speed on the on-screen display (OSD)
- `display_time_remaining`: A boolean value that indicates whether to display the time remaining (using the format HH:MM:SS) on the on-screen display (OSD)
- `font_size`: The font size for the on-screen display (OSD)
- `margin_left`: The pixel offset of the OSD from the left of the media player window
- `margin_top`: The pixel offset of the OSD from the top of the media player window
