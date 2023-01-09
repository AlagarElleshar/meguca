import {setAttrs} from "../util";

// import mpegts from "mpegts.js";

let mpegtsjs = import("mpegts.js")

export async function openFlvPlayer() {
    let mpegts = (await mpegtsjs).default
    let cont = document.getElementById("flv-player-cont")
    if (!cont) {
        cont = document.createElement("div")
        setAttrs(cont, {
            id: "flv-player-cont",
            class: "modal glass",
            style: "display: block;",
        });
        document.getElementById("modal-overlay").prepend(cont);
        if (mpegts.getFeatureList().mseLivePlayback) {
            var videoElement = document.createElement("video");
            videoElement.id = "flv-player"
            cont.appendChild(videoElement);
        }
    }
}

export async function playLive(url : string) {
    let mpegts = (await mpegtsjs).default
    if (mpegts.getFeatureList().mseLivePlayback) {
        var videoElement = document.getElementById('flv-player');
        var player = mpegts.createPlayer({
            type: 'flv',  // could also be mpegts, m2ts, flv
            isLive: true,
            url: url
        });
        player.attachMediaElement(<HTMLMediaElement>videoElement);
        player.load();
        player.play();
    }
}

(window as any).openFlvPlayer = openFlvPlayer;
(window as any).playLive = playLive;
