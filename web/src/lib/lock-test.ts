import type { WakeDocument, WakeNavigator, WakeWindow } from "./wake-lock";

// The probe of the screen lock with no tap (D-280, D-282). On the iPhone,
// the Screen Wake Lock of D-271 needs a tap after each return to the app.
// The "Screen lock test" screen runs one method at a time, and logs what
// the phone does, so the owner can pick a method:
//
//   - "wake-lock": the Wake Lock API, requested at the start and again at
//     each return, focus, and `pageshow` event, with no tap,
//   - "video-file": a silent video that the phone records from a canvas,
//     played muted in a loop, and played again at each return with no tap,
//   - "video-stream": a live silent canvas stream in a muted video, played
//     again at each return with no tap.
//
// The owner sets Auto-Lock to 30 seconds, starts a method, leaves the app,
// comes back, and does not tap. The screen must stay on for more than one
// minute. The battery use comes from the battery level before and after.

export type Method = "wake-lock" | "video-file" | "video-stream";

export const METHODS: readonly { id: Method; title: string; text: string }[] = [
  { id: "wake-lock", title: "Wake Lock, no tap", text: "The Screen Wake Lock API, asked again at each return with no tap." },
  { id: "video-file", title: "Silent video file", text: "A silent video that the phone records, played muted in a loop." },
  { id: "video-stream", title: "Silent live video", text: "A live silent stream of a canvas, played muted." },
];

// Log receives one line of the probe.
export type Log = (text: string) => void;

// The video element of a video method. A test gives a fake.
export type ProbeVideo = {
  paused: boolean;
  play: () => Promise<void>;
  pause: () => void;
  addEventListener: (type: string, fn: () => void) => void;
  removeEventListener: (type: string, fn: () => void) => void;
};

// The events of a return to the app. The document gives
// "visibilitychange", and the window gives "focus" and "pageshow".
function onReturn(doc: WakeDocument, win: WakeWindow, fn: (event: string) => void): () => void {
  const visible = () => {
    if (doc.visibilityState === "visible") fn("return");
  };
  const focus = () => fn("focus");
  const pageshow = () => fn("pageshow");
  doc.addEventListener("visibilitychange", visible);
  win.addEventListener("focus", focus);
  win.addEventListener("pageshow", pageshow);
  return () => {
    doc.removeEventListener("visibilitychange", visible);
    win.removeEventListener("focus", focus);
    win.removeEventListener("pageshow", pageshow);
  };
}

function errorName(err: unknown): string {
  return err instanceof Error ? err.name : String(err);
}

// probeWakeLock requests the lock at the start, which comes from a tap,
// and again at each return with no tap. It logs each result and each
// release, and gives the stop function.
export function probeWakeLock(nav: WakeNavigator, doc: WakeDocument, win: WakeWindow, log: Log): () => void {
  const api = nav.wakeLock;
  if (!api) {
    log("No Screen Wake Lock API.");
    return () => {};
  }
  let stopped = false;
  let held: { released: boolean; release: () => Promise<void> } | null = null;
  const request = async (event: string) => {
    if (stopped || doc.visibilityState !== "visible") return;
    if (held && !held.released) {
      log(`${event}: the lock is still on.`);
      return;
    }
    try {
      const s = await api.request("screen");
      if (stopped) {
        void s.release().catch(() => {});
        return;
      }
      held = s;
      s.addEventListener("release", () => {
        if (!stopped) log("The phone released the lock.");
      });
      log(`${event}: the lock is on.`);
    } catch (err) {
      log(`${event}: the phone refused the lock (${errorName(err)}).`);
    }
  };
  const off = onReturn(doc, win, (event) => void request(event));
  void request("start");
  return () => {
    stopped = true;
    off();
    if (held && !held.released) void held.release().catch(() => {});
  };
}

// probeVideo plays the muted video at the start, which comes from a tap.
// At each return, it logs whether the video still plays, and plays it
// again with no tap when it stopped. It gives the stop function.
export function probeVideo(video: ProbeVideo, doc: WakeDocument, win: WakeWindow, log: Log): () => void {
  let stopped = false;
  const play = async (event: string) => {
    if (stopped || doc.visibilityState !== "visible") return;
    if (!video.paused) {
      log(`${event}: the video plays.`);
      return;
    }
    try {
      await video.play();
      log(`${event}: the video plays again.`);
    } catch (err) {
      log(`${event}: the phone refused to play the video (${errorName(err)}).`);
    }
  };
  const paused = () => {
    if (!stopped) log("The video stopped.");
  };
  video.addEventListener("pause", paused);
  const off = onReturn(doc, win, (event) => void play(event));
  void play("start");
  return () => {
    stopped = true;
    off();
    video.removeEventListener("pause", paused);
    video.pause();
  };
}

// logLine gives the line of the log with the time since the start, such as
// "1:05 The video plays.".
export function logLine(start: number, at: number, text: string): string {
  const s = Math.max(0, Math.floor((at - start) / 1000));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")} ${text}`;
}

// silentCanvas gives a canvas of 16 px that changes one pixel at each
// frame, so a stream of it has frames. It gives the stop function too.
function silentCanvas(): { canvas: HTMLCanvasElement; stop: () => void } {
  const canvas = document.createElement("canvas");
  canvas.width = 16;
  canvas.height = 16;
  const ctx = canvas.getContext("2d");
  let n = 0;
  let frame = 0;
  const draw = () => {
    if (ctx) {
      ctx.fillStyle = n++ % 2 ? "#0b1220" : "#0c1321";
      ctx.fillRect(0, 0, 16, 16);
    }
    frame = requestAnimationFrame(draw);
  };
  draw();
  return { canvas, stop: () => cancelAnimationFrame(frame) };
}

// silentSource puts a silent source on the video: a file of 1 second that
// MediaRecorder records from the canvas, or the live stream of the canvas.
// It gives the stop function of the source.
export async function silentSource(video: HTMLVideoElement, method: "video-file" | "video-stream"): Promise<() => void> {
  const { canvas, stop } = silentCanvas();
  const stream = canvas.captureStream(30);
  video.muted = true;
  video.loop = true;
  video.playsInline = true;
  if (method === "video-stream") {
    video.srcObject = stream;
    return () => {
      stop();
      stream.getTracks().forEach((t) => t.stop());
      video.srcObject = null;
    };
  }
  const type = ["video/mp4", "video/webm"].find((t) => MediaRecorder.isTypeSupported(t)) ?? "";
  const recorder = new MediaRecorder(stream, type ? { mimeType: type } : undefined);
  const chunks: Blob[] = [];
  recorder.ondataavailable = (e) => chunks.push(e.data);
  const done = new Promise<void>((resolve) => (recorder.onstop = () => resolve()));
  recorder.start();
  await new Promise((r) => setTimeout(r, 1000));
  recorder.stop();
  await done;
  stop();
  stream.getTracks().forEach((t) => t.stop());
  const url = URL.createObjectURL(new Blob(chunks, { type: recorder.mimeType }));
  video.src = url;
  return () => {
    video.removeAttribute("src");
    video.load();
    URL.revokeObjectURL(url);
  };
}
