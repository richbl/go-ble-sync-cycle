<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

![bsc_example](https://github.com/user-attachments/assets/0eb908da-6d22-4e8a-a9e8-5702a01f3fe0)

At a high level, **BLE Sync Cycle** coordinates with a BLE central device (such as a computer), a BLE peripheral device (a BLE cycling sensor) and a media player (mpv or VLC), and performs the following:

### 1. Discovery and Connection

1. The BLE central device scans for the BLE peripheral device (your BLE cycling sensor)
2. The BLE central device connects to the sensor and queries for various BLE services: battery power and cycling speed

### 2. Synchronization and Real-Time Data Processing

1. The BLE central device starts receiving from the sensor real-time speed data at regular intervals

### 3. Video Playback and Display

1. The application then launches a media player for video playback
2. The application automatically adjusts video speed based on incoming cycling speed data: pedal faster and the video playback speed increases; pedal slower and the video playback speed decreases
3. The application displays real-time cycling statistics via its application interface and, optionally, the media player's on-screen display (OSD)

### 4. Application Shutdown

1. The application shuts down on user interrupt, application exit, or at the end of video playback. The shutdown process coordinates with the BLE central device, the BLE peripheral device, and the media player to ensure a smooth and clean shutdown.
