<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

### The BLE Sync Cycle Application

- <u>What is **BLE Sync Cycle**?</u>

  In its simplest form, **BLE Sync Cycle makes video playback run faster when you pedal your bike faster, and slows down video playback when you pedal slower**. And, when you stop your bike, video playback pauses.  

- <u>How do I use **BLE Sync Cycle**?</u>

  See the [Basic Usage](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-Overview) section

- <u>How do I configure **BLE Sync Cycle**?</u>

  See the [Anatomy of a BSC TOML File](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-Anatomy-of-a-BSC-TOML-File) section for an understanding of the configuration options and fields available for editing.

    - When running **BLE Sync Cycle** in GUI mode, you can edit BSC TOML configuration files using the BSC Session Editor

    - When running **BLE Sync Cycle** in CLI mode (via the command-line in a terminal), you'll need to edit these BSC TOML configuration files manually

    - > Hint: even if you decide to run **BLE Sync Cycle** in CLI mode, you can still use the BSC Session Editor (when running **BLE Sync Cycle** in GUI mode) to edit BSC TOML configuration files

- <u>Can I disable the log messages in **BLE Sync Cycle**?</u>

  The level of logging messages output in the **BLE Sync Cycle** application is configured via the `logging_level` parameter in a BSC TOML configuration file. See the [Anatomy of a BSC TOML File](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-Anatomy-of-a-BSC-TOML-File) section for details. This parameter can be set to "debug", "info", "warn", or "error", where "debug" is the most verbose (all log messages displayed), and "error" is least verbose.

- <u>My BLE sensor takes a long time to connect, and often times out. What can I do?</u>

  **This is normal**.
  
  It takes time for a BLE peripheral (your bicycle speed sensor) to first advertise its services and characteristics, and then establish a connection with a central BLE device (your laptop or computer). The **BLE Sync Cycle** application will automatically time out and notify you of the event. The easiest solution is to just restart the session (or, if you're running in CLI mode, restart **BLE Sync Cycle**), as that will usually give the BLE sensor enough time to establish a connection. If the issue persists,try increasing the `ble_connect_timeout` parameter in the `config.toml` file (see the [Anatomy of a BSC TOML File](https://github.com/richbl/go-ble-sync-cycle/wiki/Basic-Usage:-Anatomy-of-a-BSC-TOML-File) section). Different BLE devices have different advertising intervals, so you may need to adjust this value accordingly.

- <u>What videos can be used in **BLE Sync Cycle**?</u>

  **The short answer is: any video can be used**. As long as your media player is capable of playing the video file, you can use it with **BLE Sync Cycle**.

  Regarding video file formats, the only known file format that will not work with **BLE Sync Cycle** are older `mpg`/`mpeg` formats (media players will often report an invalid file format error, yet still sometimes play... odd).

  The long answer is that you will want to look for videos that are first-person cycling, driving, or even running videos. Check out [YouTube](https://www.youtube.com), [Vimeo](https://vimeo.com), [Pexels](https://www.pexels.com/videos/), or [DailyMotion](https://www.dailymotion.com/us), and search for "first person cycling" or "POV cycling" for some great ideas.

  >Hint: the next time you're planning a great outdoor cycling ride, strap on a camera and record some first-person cycling videos, and share them with this community!

### Bluetooth Protocols

- <u>Do all Bluetooth devices work with **BLE Sync Cycle**?</u>

  **No**.
  
  The Bluetooth package used by **BLE Sync Cycle**, [called Go Bluetooth by TinyGo.org](https://github.com/tinygo-org/bluetooth), is based on the [Bluetooth Low Energy (BLE) standard](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy). Note that the classic Bluetooth protocol (called Bluetooth Basic Rate/Enhanced Data Rate or BR/EDR) is not the same as the BLE protocol, so some older Bluetooth devices may not be compatible with the newer BLE protocol.

- <u>Can a single BLE peripheral device connect to multiple devices at the same time?</u>

  **Technically yes, but it's unlikely**.
  
  A Bluetooth Low Energy (BLE) network typically involves peripheral devices (like sensors, such as a Cycling Speed and Cadence, or CSC, sensor) that broadcast data, and central devices (like smartphones and computers) that connect to and receive this data. Although the BLE standard allows for the possibility of a single peripheral device connecting to multiple central devices concurrently, this feature is not commonly implemented in many commercially available BLE products.

  In practice, a typical CSC sensor like the [Magene S314 sensor](https://www.magene.com/en/all-products/60-s314-speed-cadence-dual-mode-sensor.html) will establish a connection with only one central device at a time. Therefore, if you plan to use a CSC sensor with both **BLE Sync Cycle** and a separate cycling app--like the excellent [Urban Biker](https://urban-bike-computer.com) Android app)--you will likely need to use two separate BLE sensors, each paired with its respective central device (one to the computer running **BLE Sync Cycle**, and one to your smart phone running the cycling app.
