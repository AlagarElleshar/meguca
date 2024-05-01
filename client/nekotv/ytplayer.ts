import {VideoItem} from "../typings/messages";
import {tempNotify} from "../ui/notification";
import {isNekoTVOpen, watchDiv} from "./nekotv";

const youTubeScript = document.createElement("script");
youTubeScript.src = "https://www.youtube.com/iframe_api";

export class Youtube {
    private readonly playerEl: HTMLElement = document.getElementById('#ytapiplayer');
    private player: YT.Player;
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
        }
    }

    public loadVideo(item: VideoItem): void {
        if (this.player) {
            this.player.loadVideoById(this.extractVideoId(item.url));
            return;
        }
        if (!this.isYouTubeScriptLoaded){
            this.initMediaPlayer();
        }

        this.isLoaded = false;

        this.player = new YT.Player('watch-video', {
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
                    this.player.playVideo();
                    console.log("player state", this.player.getPlayerState())
                    setTimeout(() => {
                        console.log("player state", this.player.getPlayerState())
                        if (this.player.getPlayerState() === -1) {
                            this.player.playVideo();
                            setTimeout(() => {
                                console.log("player state", this.player.getPlayerState())
                                if (this.player.getPlayerState() === -1) {
                                    tempNotify(
                                        "Click here to play this thread's live-synced video (your browser requires a click before videos can play)",
                                        "",
                                        "click-to-play",
                                        60,
                                        () => {
                                            if (this.player && this.player.playVideo) {
                                                this.player.playVideo();
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
        if (!this.player) return;

        this.isLoaded = false;
        this.player.destroy();
        this.player = null;
    }

    public isVideoLoaded(): boolean {
        return this.isLoaded;
    }

    public play(): void {
        this.player.playVideo();
    }

    public pause(): void {
        this.player.pauseVideo();
    }

    public getTime(): number {
        return this.player.getCurrentTime();
    }

    public setTime(time: number): void {
        this.player.seekTo(time, true);
    }

    public getPlaybackRate(): number {
        return this.player.getPlaybackRate();
    }

    public setPlaybackRate(rate: number): void {
        this.player.setPlaybackRate(rate);
    }
}