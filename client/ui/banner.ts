import { BannerModal } from "../base"
import {setAttrs} from "../util";
// import Mpegts from "mpegts.js";
declare let mpegts : any;
export default () => {
	for (let id of ["options", "FAQ", "identity", "account", "watcher", "flv-player"]) {
		highlightBanner(id)
	}
	new BannerModal(document.getElementById("FAQ"))
	document.getElementById("banner-flv-player").addEventListener("click", function() {
		let cont = document.getElementById("flv-player")
		if (!cont) {
			cont = document.createElement("div")
			setAttrs(cont, {
				id: "flv-player",
				class: "modal glass",
				style: "display: block;",
			});
			document.getElementById("modal-overlay").prepend(cont);
			if (mpegts.getFeatureList().mseLivePlayback) {
				var videoElement = document.createElement("video");
				cont.appendChild(videoElement);
				var player = mpegts.createPlayer({
					type: 'flv',  // could also be mpegts, m2ts, flv
					isLive: true,
					url: 'https://pull-f5-tt02-infra.fcdn.us.tiktokv.com/game/stream-3571009149474177394.flv'
				});
				player.attachMediaElement(videoElement);
				player.load();
				player.play();
			}
		}
	});

}

// Highlight options button by fading out and in, if no options are set
function highlightBanner(name: string) {
	const key = name + "_seen"
	if (localStorage.getItem(key)) {
		return
	}

	const el = document.querySelector('#banner-' + name)
	el.classList.add("blinking")

	el.addEventListener("click", () => {
		el.classList.remove("blinking")
		localStorage.setItem(key, '1')
	})
}
