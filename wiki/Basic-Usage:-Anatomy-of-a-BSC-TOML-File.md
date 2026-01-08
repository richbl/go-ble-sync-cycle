<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

The **BLE Sync Cycle** application is configured using an external configuration file named `config.toml`. A copy of that file--referred to as a "BSC TOML file"--with default values is provided in the `internal/config` directory of the project.

In general, most of the default settings found in this BSC TOML file can be left unchanged. However, here are the fields that will definitely need to be updated:

- `session_title`: A short description of the current cycling session (0-200 characters)
- `sensor_bd_addr`: The address of the BLE peripheral device (e.g., sensor) to connect with and monitor for speed data
- `wheel_circumference_mm`: The wheel circumference of the bicycle (50-3000 millimeters)
- `file_path`: The path to the video file to be played back

The default `config.toml` file is shown below:

```toml
# BLE Sync Cycle Configuration
# v0.13.0

[app]
  session_title = "Session Title" # Short description of the current cycling session (0-200 characters)
  logging_level = "info"          # Log messages generated during execution ("debug", "info", "warn", "error")

[ble]
  sensor_bd_addr = "FA:46:1D:77:C8:E1" # The Bluetooth Device Address (BD_ADDR) of the BLE peripheral
  scan_timeout_secs = 30               # Time to wait for a response from the peripheral before connect fails (1-100 seconds)

[speed]
  wheel_circumference_mm = 2155 # Wheel circumference (50-3000 millimeters)
  speed_units = "mph"           # The unit of measurement for speed ("mph" or "km/h")
  speed_threshold = 0.25        # Minimum speed change to trigger video playback update (0.00-10.00)
  smoothing_window = 5          # Number of recent speed readings to generate a stable moving average (1-25)

[video]
  media_player = "mpv"           # The video playback back-end to use ("mpv" or "vlc")
  file_path = "cycling_test.mp4" # File path to the video file for playback
  seek_to_position = "00:00"     # Starting playback position in the video ("MM:SS")
  window_scale_factor = 1.0      # Scales the size of the video window (0.1-1.0, where 1.0 = full screen)
  update_interval_secs = 0.25    # Frequency that the video player is sent speed updates (0.10-3.00 seconds)
  speed_multiplier = 0.8         # Multiplier to control video playback rate (0.1-1.5, where 0.1 = slower, 1.0 = normal, 1.5 = faster playback)

  [video.OSD]
    display_cycle_speed = true    # Display the current cycle speed on the on-screen display (true/false)
    display_playback_speed = true # Display the current video playback speed on the on-screen display (true/false)
    display_time_remaining = true # Display the current video time remaining on the on-screen display (true/false)
    font_size = 40                # Font size of the on-screen display (10-200 pixels)
    margin_left = 25              # Offset of the OSD from the left of the media player window (0-100 pixels)
    margin_top = 25               # Offset of the OSD from the top of the media player window (0-100 pixels)
```

An explanation of the various sections of the `config.toml` file is provided below:

### The App Section

The `[app]` section is used for configuration of the **BLE Sync Cycle** application itself. It includes the following parameter:

- `session_title`: A short description of the current cycling session (0-200 characters) used in the application GUI

- `logging_level`: The logging level to use, which displays messages to the console as the application executes. This can be "debug", "info", "warn", or "error", where "debug" is the most verbose and "error" is least verbose.

### The BLE Section

The `[ble]` section configures your computer (referred to as the BLE central controller) to scan for and query the BLE speed sensor (referred to as the BLE peripheral). It includes the following parameters:

- `sensor_bd_addr`: The address of the BLE peripheral device (e.g., sensor) to connect with and monitor for speed data
- `scan_timeout_secs`: The number of seconds to wait for a BLE peripheral response before generating an error message. Some BLE devices can take a while to respond (called "advertising"), so adjust this value accordingly. A value of 30 seconds is a good starting point.

> To find the address (BD_ADDR) of your BLE peripheral device, you'll need to connect to it from your computer (or any device with Bluetooth connectivity). From Ubuntu (or any other Linux distribution), you can use [the `bluetoothctl` command](https://www.mankier.com/1/bluetoothctl#). BLE peripheral device BD_ADDRs are typically in the form of "11:22:33:44:55:66."

### The Speed Section

The `[speed]` section defines the configuration for the speed controller component. The speed controller takes raw BLE CSC speed data (a rate of discrete device events per time cycle) and converts it speed (either km/h or mph, depending on `speed_units`). It includes the following parameters:

- `wheel_circumference_mm`: The wheel circumference in millimeters: important in order to accurately convert raw sensor values to actual speed (distance traveled per unit time)
- `speed_units`: The speed units to use (either "km/h" or "mph")
- `speed_threshold`: The minimum speed change to trigger video speed updates
- `smoothing_window`: The number of "look-backs" (most recent speed measurements) to use for generating a moving average for the speed value

> The smoothing window is a simple ring buffer that stores the last (n) speed measurements, meaning that it will create a moving average for the speed value. This helps to smooth out the speed data and provide a more natural video playback experience.

### The Video Section

The `[video]` section defines the configuration for the MPV video player component. It includes the following parameters:

- `media_player`: The media player to use (only "mpv" is currently supported)
- `file_path`: The full path to the video file to play. The video format must be supported by MPV (e.g., MP4, webm, etc.)
- `seek_to_position`: The minutes:seconds ("MM:SS") to seek to a specific point in video playback. This can be useful with longer videos that may take multiple training sessions to complete
- `window_scale_factor`: A scaling factor for the video window, where 1.0 is full screen. This value can be useful when debugging or when running the video player in a non-maximized window is preferred
- `update_interval_secs`: The number of seconds to wait between video player updates
- `speed_multiplier`: The relative playback speed of the video. Usually, a value of 1.0 is used (<1.0 will slow playback; >1.0 will speed up playback), as this is the default value (normal playback speed). However, since it's typically unknown what the speed of the vehicle is in the video during "normal speed" playback, it's recommended to experiment with different values to find a good balance between video playback speed and real-world cycling experience.

### The Video On-Screen Display Section

The `[video.osd]` sub-section of the `[video]` section defines the configuration for the on-screen display (OSD) functionality in the media player. It includes the following parameters:

- `display_cycle_speed`: A boolean value that indicates whether to display the cycle sensor speed on the on-screen display (OSD)
- `display_playback_speed`: A boolean value that indicates whether to display the video playback speed on the on-screen display (OSD)
- `display_time_remaining`: A boolean value that indicates whether to display the time remaining (using the format HH:MM:SS) on the on-screen display (OSD)
- `font_size`: Font size of the on-screen display (10-200 pixels)
- `margin_left`: Offset of the OSD from the left of the media player window (0-100 pixels)
- `margin_top`: Offset of the OSD from the top of the media player window (0-100 pixels)
