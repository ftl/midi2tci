## DJControl Starlight

Here is some information about my specific MIDI controller that I had trouble to find.

## Control the default behavior of the base LEDs (glowing animation)

channel 0, key 0x24: use note on/off to switch on/off

## Control the color of the base LEDs:

Use "note on" to set the color. 
Before changing the color you need to send a "note off" first.

* channel 1, key 0x23: left half
* channel 2, key 0x23: right half

The value of the velocity encodes the LED color:

* blue: bit 0 1
* green: bit 2 3 4
* red: bit 5 6 7
* white: bit 8

As bits 7..0: `wrrrgggbb`

When the "white" bit is on, the other bits are ignored.

## More information about the DJControl Starlight:
* https://github.com/mixxxdj/mixxx/wiki/Hercules%20DJ%20Control%20Starlight
* https://mixxx.discourse.group/t/hercules-djcontrol-starlight/17833
* https://mixxx.discourse.group/uploads/short-url/4D6lh9SlkjOzmmec8YZSe8wIlhX.js
