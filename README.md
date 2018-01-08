# Smart Home Samsung TV
Control your Samsung TV over the internet.

## Requirements
- A server running the [smartHomeSamsungTVServer application](https://github.com/atiernan/smartHomeSamsungTVServer)

## Example
### Search For Your TV
- Make sure your TV is switched on, and on the same network as your laptop/raspberry pi
- Run the command `smartHomeSamsungTV search` and make a note of the ID and MAC that is found

### Create a Config File
- Make a copy of the `example-config.json` file
- Replace the ID and MACAddress fields with the data discovered earlier

### Run the server
`smartHomeSamsungTVServer`

### Run the local client
`smartHomeSamsungTV -config=<Path To Your Config File>`

### Control your TV
You can now make GET requests to the server to control your TV, the list of commands are:
  - **Turn on the TV:** `http://<your server ip>/?TVOn=1`
  - **Turn off the TV:** `http://<your server ip>/?TVOff=1`
  - **Turn the TV Volume Up:** `http://<your server ip>/?VolumeUp=2` (Note you can change the query value to turn up the TV by different amounts)
  - **Turn the TV Volume Down:** `http://<your server ip>/?VolumeDown=2` (Note you can change the query value to turn down the TV by different amounts)
  - **Mute/Unmute the TV** `http://<your server ip>/?VolumeMute=1`
  - **Press Play** `http://<your server ip>/?Play=1`
  - **Press Pause** `http://<your server ip>/?Pause=1`
  - **Press OK** `http://<your server ip>/?OK=1`

You may also chain these commands together i.e. `http://<your server ip>/?TVOn=1&VolumeUp=2`

### IFTTT
You can use the webhooks service in IFTTT to control your TV, i.e. control your TV through Alexa or Google Home
- Create a new applet
- Select `Google Assistant` or `Amazon Alexa`
- Select `Say a simple phrase`, and fill out the form
- For the service select `Webhooks`
- Select `Make a web request`
- For URL enter `http://<your server ip>/?TVOn=1`
- Save the applet