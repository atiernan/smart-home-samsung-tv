# Smart Home Samsung TV
Control your Samsung TV over the internet.

## Requirements
- A server running the [smart-home-samsung-tv-server application](https://github.com/atiernan/smart-home-samsung-tv-server)
- A CEC capable HDMI port/adapter (i.e. Raspberry Pi) connected to the TV you want to control

## Example
### IFTTT
- Create a new applet
- For the trigger select `Google Assistant`
- Select `Say a simple phrase` and enter `Turn on the TV`
- For the the service select `Webhooks`
- Select `Make a web request` and for the URL select `http://<Your Server IP>/?TVOn=1`
- Save the applet

### Run the server
`tvServer`

### Run the local client
`tvClient -tv-ip=<Your TVs IP address> -server=http://<Your Server IP>/device/endpoint`