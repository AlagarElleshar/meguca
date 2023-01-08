import {setAttrs} from "../util";

import mpegts from "mpegts.js";

let player = mpegts.createPlayer({
    type: 'flv',  // could also be mpegts, m2ts, flv
    isLive: true,
});
export function openFlvPlayer() {
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
var checkBoxFields = ['isLive', 'withCredentials', 'liveBufferLatencyChasing'];
var streamURL, mediaSourceURL;

export function playLive(url : string) {
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
