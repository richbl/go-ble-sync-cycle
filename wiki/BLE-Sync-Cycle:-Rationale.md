<!-- markdownlint-disable MD033 -->
<!-- markdownlint-disable MD041 -->
<p align="center">
<img width="300" alt="BSC logo" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/ui/bsc_logo_title.png">
</p>
<!-- markdownlint-enable MD033,MD041 -->

This project was developed to address a specific need:

**How can I continue cycling when the weather outside is not ideal?**

While there are several existing solutions that allow for virtual indoor cycling, such as [Zwift](https://www.zwift.com/) and [Rouvy](https://rouvy.com/), these typically require the purchase of specialized training equipment (often precluding the use of your own bike), a subscription to compatible online virtual cycling services, and a reliable broadband Internet connection.

My needs are different:

* I want to train _using my own bicycle_. In my own case, I prefer riding recumbents, so it wouldn’t make sense for me to train on a traditional upright trainer
* I need a solution that can operate with minimal dependencies and without requiring an Internet connection, as I live in a rural part of the Pacific Northwest where both electricity and Internet services can be unreliable

> Check out my [**Watchfile Remote [Rust Edition] project**](https://github.com/richbl/rust-watchfile-remote) for an example of how I handle the notification of our regular loss of Internet service here in the woods of the Pacific Northwest

* Finally, I want flexibility in the solutions and components that I use, as I typically like to tweak the systems I work with. I suspect it's my nature as an engineer to tinker...

Since I already use an analog bicycle trainer while riding indoors, it made sense for me to find a way to pair Bluetooth cycling sensors with a local computer which could then drive some kind of interesting video feedback while cycling. This project was created to fit that need.

<p align="center">
<img width="850" alt="Screenshot showing BSC GUI" src="https://raw.githubusercontent.com/richbl/go-ble-sync-cycle/refs/heads/main/.github/assets/trainer_1_hd.png">
</p>
