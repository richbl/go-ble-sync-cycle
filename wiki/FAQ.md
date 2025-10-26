<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

### The BLE Sync Cycle Application

- <u>What is **BLE Sync Cycle**?</u>

  In its simplest form, **BLE Sync Cycle makes video playback run faster when you pedal your bike faster, and slows down video playback when you pedal slower**. And, when you stop your bike, video playback pauses.  

- <u>How do I use **BLE Sync Cycle**?</u>

  See the [Basic Usage](https://github.com/richbl/go-ble-sync-cycle/wiki/overview) section

- <u>How do I configure **BLE Sync Cycle**?</u>

  See the [Editing the TOML File](https://github.com/richbl/go-ble-sync-cycle/wiki/Editing-the-TOML-File) section

- <u>Can I disable the log messages in **BLE Sync Cycle**?</u>

  Check out the `logging_level` parameter in the `config.toml` file (see the [Editing the TOML File](https://github.com/richbl/go-ble-sync-cycle/wiki/Editing-the-TOML-File) section). This parameter can be set to "debug", "info", "warn", or "error", where "debug" is the most verbose (all log messages displayed), and "error" is least verbose.

- <u>My BLE sensor takes a long time to connect, and often times out. What can I do?</u>

  **This is normal**. It takes time for BLE sensors to first advertise their services and characteristics, and then establish a connection with a central device. The easiest solution is to just rerun **BLE Sync Cycle**, as that will usually give the BLE sensor enough time to establish a connection. If the issue persists,try increasing the `ble_connect_timeout` parameter in the `config.toml` file (see the [Editing the TOML File](https://github.com/richbl/go-ble-sync-cycle/wiki/Editing-the-TOML-File) section). Different BLE devices have different advertising intervals, so you may need to adjust this value accordingly.

- <u>What videos can be used in **BLE Sync Cycle**?</u>

  **The short answer is: any video can be used**. As long as your media player is capable of playing the video file, you can use it with **BLE Sync Cycle**.

  The long answer is that you will want to look for videos that are first-person cycling, driving, or even running videos. To get an idea for some good examples, [search YouTube for "first person cycling"](https://www.youtube.com/results?search_query=first+person+cycling).

### Bluetooth Protocols

- <u>Do all Bluetooth devices work with **BLE Sync Cycle**?</u>

  **No**. The Bluetooth package used by **BLE Sync Cycle**, [called Go Bluetooth by TinyGo.org](https://github.com/tinygo-org/bluetooth), is based on the [Bluetooth Low Energy (BLE) standard](https://en.wikipedia.org/wiki/Bluetooth_Low_Energy). Note that the classic Bluetooth protocol (called Bluetooth Basic Rate/Enhanced Data Rate or BR/EDR) is not the same as the BLE protocol, so some older Bluetooth devices may not be compatible with the newer BLE protocol.

- <u>Can a single BLE peripheral device connect to multiple devices at the same time?</u>

  **Technically yes, but it's unlikely**. A Bluetooth Low Energy (BLE) network typically involves peripheral devices (like sensors, such as a Cycling Speed and Cadence, or CSC, sensor) that broadcast data, and central devices (like smartphones and computers) that connect to and receive this data. Although the BLE standard allows for the possibility of a single peripheral device connecting to multiple central devices concurrently, this feature is not commonly implemented in many commercially available BLE products.

  In practice, a typical CSC sensor like the [Magene S314 sensor](https://www.magene.com/en/all-products/60-s314-speed-cadence-dual-mode-sensor.html) will establish a connection with only one central device at a time. Therefore, if you plan to use a CSC sensor with both **BLE Sync Cycle** and a separate cycling app (like the excellent [Urban Biker](https://urban-bike-computer.com) Android app), you will likely need to use two separate sensors, each paired with its respective central device.
