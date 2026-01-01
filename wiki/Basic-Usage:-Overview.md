<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

![bsc_example](https://github.com/user-attachments/assets/0eb908da-6d22-4e8a-a9e8-5702a01f3fe0)

At a high level, **BLE Sync Cycle** follows the steps below:

1. Scans for your BLE cycling sensor
2. Connects to the sensor and queries for various BLE services of interest: battery power and cycling speed
3. Starts receiving real-time speed data at regular intervals
4. Launches a media player for video playback
5. Automatically adjusts video speed based on your cycling speed
6. Displays real-time cycling statistics via the media player's on-screen display (OSD)
7. Gracefully shuts down on user interrupt, application exit, or at the end of video playback
