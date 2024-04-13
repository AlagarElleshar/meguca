import {importTemplate} from "../util";
import Mpegts from "mpegts.js";

// import mpegts from "mpegts.js";
// This loads mpegts.js asynchronously
let mpegtsjs = import("mpegts.js")


let playerOpen = false;
let currentURL = "";
let player: Mpegts.Player;
let playerConfig : Mpegts.Config = {
    enableWorker: true,
    liveBufferLatencyChasing: true,
    liveBufferLatencyMaxLatency: 2,
    liveBufferLatencyMinRemain: 1,
}

export function openFlvPlayer() {
    let cont = document.getElementById("flv-player-cont")
    if (!cont) {
        let playerElement = importTemplate("flv-player")
        document.getElementById("modal-overlay").prepend(playerElement);
        document.getElementById("flv-close-button").addEventListener("click", closeFlvPlayer)
        document.getElementById("flv-reload-button").addEventListener("click", reloadPlayer)
    }
    playerOpen = true
}

async function reloadPlayer() {
    player.unload()
    player.load()
    player.play()
}

function closeFlvPlayer() {
    destroyPlayer()
    let cont = document.getElementById("flv-player-cont")
    cont.remove()
    playerOpen = false
}


function destroyPlayer() {
    if (typeof player !== "undefined" && player != null) {
        player.unload();
        player.detachMediaElement();
        player.destroy();
        player = null;
    }
}

export async function playLive(url: string) {
    let mpegts = (await mpegtsjs).default
    if (mpegts.getFeatureList().mseLivePlayback) {
        var videoElement = document.getElementById('flv-player');
        destroyPlayer()
        player = mpegts.createPlayer({
            type: 'flv',  // could also be mpegts, m2ts, flv
            isLive: true,
            url: url
        }, playerConfig);
        player.attachMediaElement(<HTMLMediaElement>videoElement);
        player.load();
        player.play();
        currentURL = url;
    }
}


export async function playButtonClicked(url: string) {
    console.log("Button clicked, " + playerOpen);
    if (!playerOpen) {
        await openFlvPlayer()
    }
    await playLive(url)
}

export default function initFlvPlayer() {
    (window as any).playButtonClicked = playButtonClicked;
    (window as any).openFlvPlayer = openFlvPlayer;
    (window as any).playLive = playLive;
}

