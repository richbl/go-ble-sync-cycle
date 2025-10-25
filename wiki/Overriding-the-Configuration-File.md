<p align="center">
<picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><source media="(prefers-color-scheme: light)" srcset="https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62"><img src="[https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62](https://github.com/user-attachments/assets/12027074-e126-48d1-b9e5-25850e39dd62)" width=300></picture>
</p>

**BLE Sync Cycle** supports command-line option flags to override various configuration settings. Currently the following options are supported:

```console
 -c | --config: Path to the configuration file ('path/to/config.toml')
 -s | --seek:   Seek to a specific time in the video ('MM:SS')
 -h | --help:   Display this help message
```

## Setting the Configuration File Path

When **BLE Sync Cycle** is first started, it looks for a default configuration file called `config.toml` in the current working directory. If you want  **BLE Sync Cycle** to look in a different location, you can specify the path to the file on the command line using the `-c` (or `--config`) command line option:

```console
./ble-sync-cycle --config /path/to/config.toml
```
> Note that if you specify a configuration file using this command line option, the filename can be anything (it does not have to be `config.toml`).

### Creating a Library of BLE Sync Cycle Training Sessions

One aspect of **BLE Sync Cycle**'s ability to specify different configuration files is that you could use different files for different cycling sessions, different bicycle configurations, different sensor configurations, different videos, etc.

A training session configuration file could be created called `morning_training_italy.toml` and another called `afternoon_training_iceland.toml`, etc., each with a different set of videos and configuration settings for completely different training experiences. To start such a training session, you would run the following:

```console
./ble-sync-cycle --config /path/to/afternoon_training_iceland.toml
```

## Seeking to a Specific Time in the Video

If you want to seek to a specific time in the video (useful in particularly long videos), you can use the `-s` (or `--seek`) command line option. For example, to seek to 10 minutes and 30 seconds into the video, you would use the following command:

```console
./ble-sync-cycle --seek 10:30
```

## Displaying Help in **BLE Sync Cycle**

To display the help message, you can use the `-h` (or `--help`) command line option.

```console
./ble-sync-cycle --help
```

The output of that command is as follows:

```console
 -c | --config: Path to the configuration file ('path/to/config.toml')
 -s | --seek:   Seek to a specific time in the video ('MM:SS')
 -h | --help:   Display this help message
```
