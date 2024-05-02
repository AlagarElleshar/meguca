import {VideoItem} from "../typings/messages";
import {tempNotify} from "../ui/notification";
import {isNekoTVOpen, watchDiv} from "./nekotv";
import options from "../options";

const youTubeScript = document.createElement("script");
youTubeScript.src = "https://www.youtube.com/iframe_api";
export let ytPlayer: YT.Player;

export class Youtube {
    private readonly playerEl: HTMLElement = document.getElementById('#ytapiplayer');
    private isLoaded = false;
    private isYouTubeScriptLoaded: boolean;

    public isSupportedLink(url: string): boolean {
        return this.extractVideoId(url) !== '';
    }

    public extractVideoId(url: string): string {
        const matchId = /youtube\.com.*v=([A-z0-9_-]+)/;
        const matchShort = /youtu\.be\/([A-z0-9_-]+)/;
        const matchShorts = /youtube\.com\/shorts\/([A-z0-9_-]+)/;
        const matchEmbed = /youtube\.com\/embed\/([A-z0-9_-]+)/;

        if (matchId.test(url)) {
            return url.match(matchId)[1];
        }
        if (matchShort.test(url)) {
            return url.match(matchShort)[1];
        }
        if (matchShorts.test(url)) {
            return url.match(matchShorts)[1];
        }
        if (matchEmbed.test(url)) {
            return url.match(matchEmbed)[1];
        }
        return '';
    }
    public initMediaPlayer() {
        if (!this.isYouTubeScriptLoaded && isNekoTVOpen()) {
            watchDiv.classList.remove("hidden");
            document.head.appendChild(youTubeScript);
            console.log("Load YouTube player script");
            this.isYouTubeScriptLoaded = true;
            options.onChange("audioVolume", (vol) => {
                this.setPlayerVolume();
            });
        }
    }

    public loadVideo(item: VideoItem): void {
        if (ytPlayer) {
            ytPlayer.loadVideoById(this.extractVideoId(item.url));
            return;
        }
        if (!this.isYouTubeScriptLoaded){
            this.initMediaPlayer();
        }

        this.isLoaded = false;

        ytPlayer = new YT.Player('watch-video', {
            videoId: this.extractVideoId(item.url),
            playerVars: {
                autoplay: 1,
                playsinline: 1,
                modestbranding: 1,
                rel: 0,
                showinfo: 0,
            },
            events: {
                onReady: () => {
                    this.isLoaded = true;
                    this.setPlayerVolume()
                    ytPlayer.playVideo();
                    console.log("player state", ytPlayer.getPlayerState())
                    setTimeout(() => {
                        console.log("player state", ytPlayer.getPlayerState())
                        if (ytPlayer.getPlayerState() === -1) {
                            ytPlayer.playVideo();
                            setTimeout(() => {
                                console.log("player state", ytPlayer.getPlayerState())
                                if (ytPlayer.getPlayerState() === -1) {
                                    tempNotify(
                                        "Click here to play this thread's live-synced video (your browser requires a click before videos can play)",
                                        "",
                                        "click-to-play",
                                        60,
                                        () => {
                                            if (ytPlayer && ytPlayer.playVideo) {
                                                ytPlayer.playVideo();
                                            }
                                        }
                                    );
                                }
                            }, 1000);
                        }
                    }, 2000);
                },
            },
        });
    }

    public removeVideo(): void {
        if (!ytPlayer) return;

        this.isLoaded = false;
        ytPlayer.destroy();
        ytPlayer = null;
    }

    public isVideoLoaded(): boolean {
        return this.isLoaded;
    }

    public play(): void {
        ytPlayer.playVideo();
    }

    public pause(): void {
        ytPlayer.pauseVideo();
    }

    public getTime(): number {
        return ytPlayer.getCurrentTime();
    }

    public setTime(time: number): void {
        ytPlayer.seekTo(time, true);
    }

    public getPlaybackRate(): number {
        return ytPlayer.getPlaybackRate();
    }

    public setPlaybackRate(rate: number): void {
        ytPlayer.setPlaybackRate(rate);
    }

    public setPlayerVolume(volume:number=null) {
        if (ytPlayer && ytPlayer.setVolume) {
            if(!volume){
                volume = options.audioVolume;
            }
            ytPlayer.setVolume(volume);
        }
    }

    getVideo() {

    }
}