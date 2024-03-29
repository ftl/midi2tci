# midi2tci

midi2tci allows you to control the ExpertSDR software through the TCI protocol using an USB MIDI device.

This tool is written in Go on Linux, but it works also on Windows and OSX.

## Usage

To run the tool you just have to provide the name of your configuration file:

```
$ midi2tci --config ./example_config.json
```

The configuration file contains the mappings of MIDI input controls to TCI commands and all other required settings ([see below](#setup)).

## Setup

Putting together the configuration file is done in two steps: first you need to find out on which MIDI port your device is connected, then you have to find out what channel and key your desired MIDI input controls are using. [example_config.json](./example_config.json) contains an example configuration with all available functions, which are documented also the [wiki](https://github.com/ftl/midi2tci/wiki/Functions).

### Find the MIDI Device

The tool shows you all available MIDI devices with the `list` command:

```
$ midi2tci list

2021/08/08 09:55:33 available input devices:
2021/08/08 09:55:33  0: Midi Through:Midi Through Port-0 14:0
2021/08/08 09:55:33  1: DJControl Starlight:DJControl Starlight MIDI 1 24:0
2021/08/08 09:55:33 
2021/08/08 09:55:33 available output devices:
2021/08/08 09:55:33  0: Midi Through:Midi Through Port-0 14:0
2021/08/08 09:55:33  1: DJControl Starlight:DJControl Starlight MIDI 1 24:0

```

Here you can see that my "DJControl Starlight" controller is connected on MIDI port 1.

### Find the MIDI Mappings

Each MIDI input control has an individual channel and key setting. To find you the settings of a specific control, you can use the `--trace` parameter:

```
$ midi2tci --portNumber=1 --trace

2021/08/08 09:59:50 Cannot read configuration file: open ./config.json: no such file or directory
2021/08/08 09:59:50 Opened DJControl Starlight:DJControl Starlight MIDI 1 24:0 successfully for writing
2021/08/08 09:59:50 Opened DJControl Starlight:DJControl Starlight MIDI 1 24:0 successfully for reading
2021/08/08 09:59:51 rx: channel.ControlChange{channel:0x1, controller:0xa, value:0x1}
```

The example shows that the control I want to use for VFOA has the setting channel=1 and key=10 (= 0x0a). 

```
2021/08/08 09:59:59 rx: channel.Pitchbend{channel:0x0, value:4304, absValue:0x30d0}
```

This example shows a control that send Pitchbend events over MIDI. In this case the key parameter of the mapping needs to be -1 (key=-1).

## Supported Hardware

The general behavior of MIDI controllers should be independent of the actual hardware that you are using. So any MIDI controller should work in theory. However, the devil lies in the details, and your specific hardware might behave differently. I tested the following hardware:

* Hercules DJControl Starlight
* Behringer CMD PL-1 (kudos to Elmar/DG7YEO)
* NumarkDJ2GO2 Touch  

I found some differences in the way that the LED indicators are controlled. Therefor there is the parameter `indicators` in the configuration file, that currently accepts the following values:

* default: the default, based on the behavior of the Hercules DJControl Starlight
* pl-1: special treatment of LED indicators for rotary encoders as the Behringer CMD PL-1 expects it

## License

This tool is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)
