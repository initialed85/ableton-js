import { Ableton } from "./index";

const readline = require("readline");

const ableton = new Ableton();

const setup = async () => {
  await ableton.song.addListener("is_playing", (p) => {
    console.log(`is_playing: ${p}; ping: ${ableton.getPing()}`);
  });
  await ableton.song.addListener("tempo", (t) => {
    console.log(`tempo: ${t}; ping: ${ableton.getPing()}`);
  });
  await ableton.song.addListener("current_song_time", (c) => {
    console.log(`current_song_time: ${c}; ping: ${ableton.getPing()}`);
  });

  const tempo = await ableton.song.get("tempo");
  console.log(`tempo: ${tempo}; ping: ${ableton.getPing()}`);
};

setup().then(() => {
  // noop
});

process.on("SIGINT", function () {
  console.log("Caught interrupt signal");

  try {
    ableton.close();
  } catch (e) {
    // noop
  }
});

const keypress = async () => {
  process.stdin.setRawMode(true);
  return new Promise((resolve) =>
    process.stdin.once("data", () => {
      process.stdin.setRawMode(false);
      resolve(true);
    }),
  );
};

const doSomethingOnKeyPress = async () => {
  while (true) {
    await keypress();
    await ableton.song.get("metronome");
  }
};

doSomethingOnKeyPress().then(() => {
  // noop
});
