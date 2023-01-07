import {setAttrs} from "../util";

// declare let mpegts : any;
import Mpegts from "mpegts.js";

let player = Mpegts.createPlayer({
    type: 'flv',  // could also be Mpegts, m2ts, flv
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
        if (Mpegts.getFeatureList().mseLivePlayback) {
            var videoElement = document.createElement("video");
            videoElement.id = "flv-player"
            cont.appendChild(videoElement);
        }
    }
}
var checkBoxFields = ['isLive', 'withCredentials', 'liveBufferLatencyChasing'];
var streamURL, mediaSourceURL;

export function playLive(url : string) {
    if (Mpegts.getFeatureList().mseLivePlayback) {
        var videoElement = document.getElementById('flv-player');
        var player = Mpegts.createPlayer({
            type: 'flv',  // could also be Mpegts, m2ts, flv
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
