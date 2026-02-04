<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>

<!-- markdownlint-disable MD033 -->
<p align="center">
<img width="850" alt="Screenshot showing cycling trainer" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/wa_mountain_trail_hd.png">
</p>
<!-- markdownlint-enable MD033 -->

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

## A More Technical View of the Application Workflow

For those with an interest in how the various controllers and services work collaboratively in **BLE Sync Cycle**, here's a sequence diagram for the "happy path" use case, where a user selects and loads a session, starts a session, and "cycles" the session until completion.

```mermaid
---
config:
  theme: 'forest'
---
sequenceDiagram
    rect rgb(230, 230, 230)
    participant User
    participant UI
    participant SM as SessionManager
    participant SD as ShutdownManager
    participant BLE as BLEController
    participant VID as VideoController

    User->>UI: Click Start Session
    UI->>SM: StartSession()
    
    rect rgb(240, 248, 255)
    note right of SM: Initialization Phase
    SM->>SM: Set State = Connecting
    SM->>SD: NewShutdownManager()
    SM->>SM: InitializeControllers()
    SM->>BLE: NewBLEController()
    SM->>VID: NewPlaybackController()
    end

    rect rgb(255, 248, 240)
    note right of SM: Connection Phase
    SM->>SM: Set State = Connected
    SM->>BLE: ScanForBLEPeripheral()
    activate BLE
    BLE-->>SM: ScanResult (Found)
    deactivate BLE
    
    SM->>BLE: ConnectToBLEPeripheral()
    activate BLE
    BLE-->>SM: Device Connected
    deactivate BLE
    
    SM->>BLE: DiscoverServices (Battery/CSC)
    end

    rect rgb(240, 255, 240)
    note right of SM: Runtime Phase
    SM->>SM: Set State = Running
    SM->>SD: Run(BLEUpdates)
    SM->>SD: Run(StartPlayback)
    end
    
    SM-->>UI: Return nil (Success)
    UI->>UI: Start Metrics Loop
    end
```
