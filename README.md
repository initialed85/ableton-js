# More of Ed's hacking

If you wanna run the Go stuff- spin up all the usual stuff (except for `scratch.ts`) and then run the following:

```shell
go mod download # not 100% sure you need this, go run might take care of it

go run cmd/main.go
```

You should end up with output like this:

```
2022/03/27 13:04:26 93.923603ms to get {"event": "result", "data": 3644.5013151927437, "uuid": "4d58f6c8-0c18-4c78-a993-fc468ba4baca"}
2022/03/27 13:04:26 93.489723ms to get {"event": "result", "data": 3644.681315192744, "uuid": "e036f213-b347-45b9-9bcd-c7a528eeca55"}
2022/03/27 13:04:26 103.732825ms to get {"event": "result", "data": 3644.921315192744, "uuid": "bcde1aa8-acc7-4b69-82bd-7d07118ec8fa"}
2022/03/27 13:04:26 1.454081ms to get {"event": "result", "data": 3644.921315192744, "uuid": "3af6c369-8109-4c30-8068-670ceaf5b84b"}
2022/03/27 13:04:26 2.52519ms to get {"event": "result", "data": 3645.1013151927436, "uuid": "4398c0da-3b9b-4782-90c9-7e4a72a4608a"}
2022/03/27 13:04:26 4.287073ms to get {"event": "result", "data": 3645.321315192744, "uuid": "366f189f-71e9-46c0-80c6-edca3200eba9"}
2022/03/27 13:04:26 3.5355ms to get {"event": "result", "data": 3645.5213151927437, "uuid": "75868119-85d8-48f1-9fcd-20c1490c6502"}
2022/03/27 13:04:27 4.392925ms to get {"event": "result", "data": 3645.701315192744, "uuid": "b76f42bd-bc28-44b2-9f84-3453d623bf15"}
2022/03/27 13:04:27 4.241576ms to get {"event": "result", "data": 3645.921315192744, "uuid": "fc78f4c7-25a7-4c29-9886-db1a4e0bba72"}
2022/03/27 13:04:27 4.183697ms to get {"event": "result", "data": 3646.1013151927436, "uuid": "c139abe0-dfcf-4ea9-9b1d-e39e9f2fc429"}
2022/03/27 13:04:27 13.166169ms to get {"event": "result", "data": 3646.341315192744, "uuid": "f127c39d-832e-4bdb-b895-b37b7217093d"}
2022/03/27 13:04:27 14.839087ms to get {"event": "result", "data": 3646.5213151927437, "uuid": "2a6a8265-af81-4289-9111-bc824223e7af"}
2022/03/27 13:04:27 24.091779ms to get {"event": "result", "data": 3646.761315192744, "uuid": "442783d1-890c-4d77-bc1d-2d8bd5fab3b1"}
2022/03/27 13:04:27 33.608012ms to get {"event": "result", "data": 3646.9613151927438, "uuid": "e6ed0368-eb4b-4fa6-be1a-c8cea969a09b"}
2022/03/27 13:04:27 34.362371ms to get {"event": "result", "data": 3647.181315192744, "uuid": "c9fdcdee-0fa8-447c-8f29-61268b0e959d"}
2022/03/27 13:04:27 43.55298ms to get {"event": "result", "data": 3647.381315192744, "uuid": "9f5f9233-16df-45b8-94eb-663356523fbd"}
2022/03/27 13:04:27 43.943875ms to get {"event": "result", "data": 3647.5813151927437, "uuid": "47cc0e52-8e9b-4f5e-b2bc-05041dad8dc0"}
```

As we can see, there's still this odd latency / jitter- I don't _think_ it's to do with my Go code, not impossible of course.

Current theory relates to the Python GIL (Global Interpreter Lock)- basically a lock that gets held whenever a thread has to do any sort
of memory access (maybe just whenever a thread is executing) and in particular it gets interacted with when there are syscalls at play
(e.g. networking stuff).

I don't know enough about how Ableton is actually running a Python interpreter as part of the application, perhaps it does some careful
GIL stuff (as one of the Reddit users suggested) and / or perhaps Ableton just has a lot of Python threads running in the background in
the same interpreter (which will of course all need to use the GIL for safety).

It may even be that they lock the GIL any time they update something behind the API (which would make sense, risk of memory / thread safety
issues otherwise).

I'll have a think about what else we could do, but we might be in a rock and a hard place here- _possibly_ something could be done
with Python's multiprocessing (i.e. we could fork and handle the network IO in the forked subprocess and then use a multiprocesisng queue
to share data), but I don't know for sure that multiprocessing interactions won't also count as syscalls (and therefore have the same 
issues).

# Ed's hacking

All I've really done is add some ghetto profiling to understanding how long a few different things take and then because I've got NFI what
I'm doing when it comes to Node.js I found some thing on Stackoverflow that said to use `ts-node` to just kind of handle the whole
transpile-and-run piece for you and put together `scratch.ts` to interact w/ the UDP server in the control surface on the Live side.

So, to test what I've been tinkering with:

**Shell 1**

```shell
# spin up Live (you'll need to set up the control surface)
yarn ableton11:start
```

**Shell 2**

```shell
# this should install ts-node 
yarn install

# this fires off a few subscriptions and then runs a command any time you hit a key (and spews some timing output)
npx ts-node src/scratch.ts
```

## Findings

I think you've actually done a good job of the Python UDP server, as far as I can tell it's not responsible for any significant latency (at
least not with my contrived testing).

You'll notice that you see things like this on the Node.js side (using `scratch.ts):

```
sendCommand; 1 ms
getPing; 14 ms
handleIncoming; 0 ms
getProp; 15 ms
get; 15 ms
```

I didn't put much timing effort any further in because it's pretty complex w/ the whole message map and it's late here.

At any rate, it definitely supports that there is a latency / jitter problem (sometimes it's like 90ms).

Over on the Python side though, we see stuff like this (in the Live logs):

```
2022-03-27T00:40:33.904438: info: RemoteScriptMessage: (AbletonJS) Receiving: {'uuid': 'ef84c148-8a4a-4b5c-ac2e-97447e744db4', 'ns': 'song', 'name': 'get_prop', 'args': {'prop': 'current_song_time'}}
2022-03-27T00:40:33.904596: info: RemoteScriptMessage: (AbletonJS) Received command: {'uuid': 'ef84c148-8a4a-4b5c-ac2e-97447e744db4', 'ns': 'song', 'name': 'get_prop', 'args': {'prop': 'current_song_time'}}
2022-03-27T00:40:33.904826: info: RemoteScriptMessage: Interface.handle took 0.048 ms to invoke get_prop
2022-03-27T00:40:33.905190: info: RemoteScriptMessage: (AbletonJS) Socket.send took 0.129 ms to send result
2022-03-27T00:40:33.905327: info: RemoteScriptMessage: Interface.handle took 0.545 ms to handle get_prop
2022-03-27T00:40:33.905635: info: RemoteScriptMessage: (AbletonJS) AbletonJS.command_handler took 0.7789999999999999 ms to handle {'uuid': 'ef84c148-8a4a-4b5c-ac2e-97447e744db4', 'ns': 'song', 'name': 'get_prop', 'args': {'prop': 'current_song_time'}}
2022-03-27T00:40:33.905842: info: RemoteScriptMessage: (AbletonJS) Socket.process took 1.589 ms for happy path loop iteration
```

You can see we're spending very little time overall- handle the request, do the interactions w/ Live's API, fire back a response in <2ms.

I don't really have a smoking gun, but I think we're probably seeing some latency somewhere in the way Node.js does IO (maybe async
funniness, not sure)- not very scientific, I think the next thing I'll try (probably tomorrow night) is reimplementing a very minimal
version of the client in another language (probably Golang because I'm comfortable with it and it's fast) to see if the latency problem goes
away.

If it doesn't go away, then perhaps the issue could still be Python, but maybe something to do with when buffers are flushed after you call
`socket.send`.

I don't _think_ that's likely though, I recall I've implemented a low latency thing (Python to Python) using UDP at least once for the
purposes of controlling an RC car using a PS4 controller (so that included a WiFi hop and even then it was pretty responsive).

Code for that is [here](https://github.com/initialed85/pi-rc-car/tree/master/phase_3) for reference, but it's not a good example of
anything.

## Unrelated suggestions

If you find once your UDP server gets really busy (lots of back and forth from the Node.js side to the Python side) then it may be that the
single-threaded (i.e. read -> handle -> respond) approach might be a bottleneck- especially if some requests against the Live API take
longer than others.

In this case, you could look at passing off the "handle" part of the process to another thread, however you would probably need to protect
interactions with your socket using a lock (like a mutex) because I don't think Python sockets are thread safe (and so if you're reading in
one thread and writing from potentially many other handler worker threads you might end up breaking things).

That might be okay or it might not- you might end up slowing yourself down again (because basically the lock will ensure you're never
reading and writing simultaneously).

If you plan for this thing to get really, really busy it might be better to separate out your request and response channels (architecturally
speaking)- that is to say have one socket for requests and one socket for responses (and naturally the same on the Node.js client side).

Then your process might look a bit like this

- Python side boots up
    - have a thread w/ a loop that only reads requests from the request socket and gives them to a pool of handler threads
    - your pool of handler threads (e.g. ThreadPoolExecutor) are responsible for taking a single request, making the single Live API call
      and putting the response in a queue
    - have a thread w/ a loop that only reads responses from the queue and sends them down the response socket

Or something to that effect- basically the theme is get your request socket back to the listening state as soon as possible and do your Live
API work as concurrently (and therefore quickly, wall-clock-wise) as possible; the response socket is used so that the process of responding
doesn't get in the way of handling requests.

A possible gotcha here is if the Live API is not thread safe- then you won't want a thread pool for handling Live API calls, instead you'll
just want a single thread probably fed from a queue (again, you still wanna get requests out of the way of the request socket quickly, but
you'll be limited to processing them against the Live API in a serial fashion).

Anyway- hope this is helpful!

# ---- ---- ---- ---- Original README ---- ---- ---- ----

# Ableton.js

[![Current Version](https://img.shields.io/npm/v/ableton-js.svg)](https://www.npmjs.com/package/ableton-js/)

Ableton.js lets you control your instance or instances of Ableton using Node.js. It tries to cover as many functions as possible.

This package is still a work-in-progress. My goal is to expose all of
[Ableton's MIDI Remote Script](https://julienbayle.studio/PythonLiveAPI_documentation/Live10.0.2.xml)
functions to TypeScript. If you'd like to contribute, please feel free to do so.

## Sponsored Message

I've used Ableton.js to build a setlist manager called
[AbleSet](https://ableset.app). AbleSet allows you to easily manage and control your Ableton setlists from any device, re-order songs and
add notes to them, and get an overview of the current of your set.

[![AbleSet Header](https://public-files.gumroad.com/variants/oplxt68bsgq1hu61t8bydfkgppr5/baaca0eb0e33dc4f9d45910b8c86623f0144cea0fe0c2093c546d17d535752eb)](https://ableset.app)

## Prerequisites

To use this library, you'll need to install and activate the MIDI Remote Script in Ableton.js. To do that, copy the `midi-script` folder of
this repo to Ableton's Remote Scripts folder and rename it to `AbletonJS`. The MIDI Remote Scripts folder is usually located at:

- **Windows:** {path to Ableton}\Resources\MIDI\Remote Scripts
- **macOS:** /Applications/Ableton Live {version}/Contents/App-Resources/MIDI Remote Scripts

After starting Ableton Live, add the script to your list of control surfaces:

![Ableton Live Settings](https://i.imgur.com/a34zJca.png)

If you've forked this project on macOS, you can also use yarn to do that for you. Running `yarn ableton:start` will copy the `midi-script`
folder, open Ableton and show a stream of log messages until you kill it.

## Using Ableton.js

This library exposes an `Ableton` class which lets you control the entire application. You can instantiate it once and use TS to explore
available features.

Example:

```typescript
import {Ableton} from "ableton-js";

const ableton = new Ableton();

const test = async () => {
    ableton.song.addListener("is_playing", (p) => console.log("Playing:", p));
    ableton.song.addListener("tempo", (t) => console.log("Tempo:", t));

    const tempo = await ableton.song.get("tempo");
    console.log(tempo);
};

test();
```

## Events

There are a few events you can use to get more under-the-hood insights:

```ts
// A connection to Ableton is established
ab.on("connect", (e) => console.log("Connect", e));

// Connection to Ableton was lost,
// also happens when you load a new project
ab.on("disconnect", (e) => console.log("Disconnect", e));

// A raw message was received from Ableton
ab.on("message", (m) => console.log("Message:", m));

// A received message could not be parsed
ab.on("error", (e) => console.error("Error:", e));

// Fires on every response with the current ping
ab.on("ping", (ping) => console.log("Ping:", ping, "ms"));
```

## Protocol

Ableton.js uses UDP to communicate with the MIDI Script. Each message is a JSON object containing required data and a UUID so request and
response can be associated with each other.

### Compression and Chunking

To allow sending large JSON payloads, requests to and responses from the MIDI Script are compressed using gzip and chunked every 7500 bytes.
The first byte of every message contains the chunk index (0x00-0xFF) followed by the gzipped chunk. The last chunk always has the index
0xFF. This indicates to the JS library that the previous received messages should be stiched together, unzipped, and processed.

### Commands

A command payload consists of the following properties:

```js
{
    "uuid"
:
    "a20f25a0-83e2-11e9-bbe1-bd3a580ef903", // A unique command id
        "ns"
:
    "song", // The command namespace
        "nsid"
:
    null, // The namespace id, for example to address a specific track or device
        "name"
:
    "get_prop", // Command name
        "args"
:
    {
        "prop"
    :
        "current_song_time"
    } // Command arguments
}
```

The MIDI Script answers with a JSON object looking like this:

```js
{
    "data"
:
    0.0, // The command's return value, can be of any JSON-compatible type
        "event"
:
    "result", // This can be 'result' or 'error'
        "uuid"
:
    "a20f25a0-83e2-11e9-bbe1-bd3a580ef903"
}
```

### Events

To attach an event listener to a specific property, the client sends a command object:

```js
{
    "uuid"
:
    "922d54d0-83e3-11e9-ba7c-917478f8b91b", // A unique command id
        "ns"
:
    "song", // The command namespace
        "name"
:
    "add_listener", // The command to add an event listener
        "args"
:
    {
        "prop"
    :
        "current_song_time", // The property that should be watched
            "eventId"
    :
        "922d2dc0-83e3-11e9-ba7c-917478f8b91b" // A unique event id
    }
}
```

The MIDI Script answers with a JSON object looking like this to confirm that the listener has been attached:

```js
{
    "data"
:
    "922d2dc0-83e3-11e9-ba7c-917478f8b91b", // The unique event id
        "event"
:
    "result", // Should be result, is error when something goes wrong
        "uuid"
:
    "922d54d0-83e3-11e9-ba7c-917478f8b91b" // The unique command id
}
```

From now on, when the observed property changes, the MIDI Script sends an event object:

```js
{
    "data"
:
    68.0, // The new value, can be any JSON-compatible type
        "event"
:
    "922d2dc0-83e3-11e9-ba7c-917478f8b91b", // The event id
        "uuid"
:
    null // Is always null and may be removed in future versions
}
```

Note that for some values, this event is emitted multiple times per second. 20-30 updates per second are not unusual.

### Connection Events

The MIDI Script sends events when it starts and when it shuts down. These look like this:

```js
{
    "data"
:
    null, // Is always null
        "event"
:
    "connect", // Can be connect or disconnect
        "uuid"
:
    null // Is always null and may be removed in future versions
}
```

When you open a new Project in Ableton, the script will shut down and start again.

When Ableton.js receives a disconnect event, it clears all current event listeners and pending commands. It is usually a good idea to attach
all event listeners and get properties each time the `connect` event is emitted.

### Findings

In this section, I'll note interesting pieces of information related to Ableton's Python framework that I stumble upon during the
development of this library.

- It seems like Ableton's listener to `output_meter_level` doesn't quite work as well as expected, hanging every few 100ms. Listening
  to `output_meter_left` or
  `output_meter_right` works better. See
  [Issue #4](https://github.com/leolabs/ableton-js/issues/4)
- The `playing_status` listener of clip slots never fires in Ableton. See
  [Issue #25](https://github.com/leolabs/ableton-js/issues/25)
