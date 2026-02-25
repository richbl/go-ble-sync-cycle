<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

In addition to having a standard Go environment installed and operational,the following additional libraries are required in order to successfully compile/build the **BLE Sync Cycle** application:

### The GObject Introspection Package

- This library used by Go provides machine readable introspection data of the API of C libraries, necessary as the GTK4/Adwaita libraries are natively C-based. To install:

  ``` console
  sudo apt install libgirepository1.0-dev 
  ```

### The GTK4 Development Library

- [GTK4](https://docs.gtk.org/gtk4/index.html) is an open source library for designing graphical user interfaces (GUIs) for the Gnome desktop environment on Linux (also available on macOS and Windows). In order to use this library (necessary for building the **BLE Sync Cycle** application), it needs to be installed locally. To do so:

  ```console
  sudo apt install libgtk-4-dev
  ```

### The Adwaita (libadwaita) Development Library

- As part of the GTK4 development library, [Adwaita](https://gnome.pages.gitlab.gnome.org/libadwaita/) is an extension used to meet the guidance of the [Gnome Human Interface Guidelines (HIG)](https://developer.gnome.org/hig/). To install this library:

  ```console
  sudo apt install libadwaita-1-dev
  ```

### The mpv Media Player Library

- **BLE Sync Cycle** relies on the [mpv](https://mpv.io/) media player library for video playback. To install this library:

    ```console
    sudo apt-get install libmpv2
    ```

> Note that the [mpv](https://mpv.io/) media player itself does not need to be installed (just the mpv library)
