# Lucas' Light Alarm

Webserver for hosting a light alarm on a raspberry pi. Expects
the WS2812B compatible lights to have an external power source
and for data to be plugged into pin 18.

Build:
go build -o lightalarm

Run:
./lightalarm
or
go run main.go


Daemon:
It's currently set up as a systemd daemon defined in `/etc/systemd/system/light-alarm.service`
It currently runs in `home/pi/light_alarm/lightalarm`

To restart - `sudo systemctl restart light-alarm.service`