<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->
The **BLE Sync Cycle** project currently only offers the option for the local compilation/build of the application. While future releases may provide pre-built binaries (or a flatpak release), the following software requirements are needed:

### A Go Environment

- In order to compile the executable for this project, an operational [Go language](https://go.dev/) environment is required (with several additional libraries installed: see [Installation:-Application-Dependencies](https://github.com/richbl/go-ble-sync-cycle/wiki/Installation:-Application-Dependencies) for details).

> Once the **BLE Sync Cycle** application is compiled into an executable, it can be run without further dependencies on Go

### A Media Player

- The open source, cross-platform [mpv media player](https://mpv.io/), installed (e.g., `sudo apt-get install mpv`) and operational

  OR

- The open source, cross-platform [VLC media player](https://www.videolan.org/vlc), installed (e.g., `sudo apt-get install vlc`) and operational

> It's highly recommended to use the [mpv media player](https://mpv.io/) whenever possible, as the [VLC media player](https://www.videolan.org/vlc) can cause application instability if a user closes the VLC playback window while a **BLE Sync Cycle** session is running. This is a known issue.

### A First-Person View Cycling Video

- A local video file for playback, preferably a first-person view cycling video.  Check out [YouTube](https://www.youtube.com), [Vimeo](https://vimeo.com), [Pexels](https://www.pexels.com/videos/), or [DailyMotion](https://www.dailymotion.com/us), and search for "first person cycling" or "POV cycling" for some great ideas

### The Target Platform

While **BLE Sync Cycle** has been written and tested using Ubuntu 24.04 through 25.10 on AMD and Intel processors, it should work across any comparable Unix-like platform and architecture
