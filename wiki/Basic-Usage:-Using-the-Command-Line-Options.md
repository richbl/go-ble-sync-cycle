<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->
When running in CLI mode, **BLE Sync Cycle** supports several command-line option flags to override various configuration settings. These flags are:

```console
-----------------------------------------------------------------------------------

Usage: ble-sync-cycle [flags]

The following flags are available when running in console/CLI mode:

  -n, --no-gui       Run the application without a graphical user interface (GUI)
  -c, --config       Path to the configuration file ('path/to/config.toml')
  -s, --seek         Seek to a specific time in the video ('MM:SS')
  -h, --help         Display this help message

The following flag is available when running in GUI mode:

  -l, --log-console  Enable logging to the console

-----------------------------------------------------------------------------------

```

### Running **BLE Sync Cycle** in CLI Mode

To run **BLE Sync Cycle** in CLI mode, you can use the `-n` (or `--no-gui`) command line option:

```console
./ble-sync-cycle --no-gui
```

### Setting the Configuration File Path

When **BLE Sync Cycle** is first started in CLI mode, it looks for a default configuration file called `config.toml` in the current working directory. If you want  **BLE Sync Cycle** to look in a different location, you can specify the path and filename of the configuration file using the `-c` (or `--config`) command line option:

```console
./ble-sync-cycle --no-gui --config /path/to/my-bsc-config.toml
```

> Note that if you specify a configuration file using this command line option, the filename can be anything (it does not have to be `config.toml`).

#### Creating a Library of BLE Sync Cycle Training Sessions

One aspect of **BLE Sync Cycle**'s ability to specify different configuration files is that you could use different files for different cycling sessions, different bicycle configurations, different sensor configurations, different videos, etc.

A training session configuration file could be created called `morning_training_italy.toml` and another called `afternoon_training_iceland.toml`, etc., each with a different set of videos and configuration settings for completely different training experiences. To start such a training session, you would run the following:

```console
./ble-sync-cycle --no-gui --config /path/to/morning_training_italy.toml
```

And then later in the day you would run:

```console
./ble-sync-cycle --no-gui --config /path/to/afternoon_training_iceland.toml
```

### Seeking to a Specific Time in the Video

If you want to seek to a specific time in the video (useful in particularly long videos), you can use the `-s` (or `--seek`) command line option. For example, to seek to 10 minutes and 30 seconds into the video, you would use the following command:

```console
./ble-sync-cycle --no-gui --seek 10:30
```

### Displaying Help in **BLE Sync Cycle**

To display the help message, you can use the `-h` (or `--help`) command line option.

```console
./ble-sync-cycle --help
```

The output of that command is as follows:

```console
-----------------------------------------------------------------------------------

Usage: ble-sync-cycle [flags]

The following flags are available when running in console/CLI mode:

  -n, --no-gui       Run the application without a graphical user interface (GUI)
  -c, --config       Path to the configuration file ('path/to/config.toml')
  -s, --seek         Seek to a specific time in the video ('MM:SS')
  -h, --help         Display this help message

The following flag is available when running in GUI mode:

  -l, --log-console  Enable logging to the console

-----------------------------------------------------------------------------------
```
