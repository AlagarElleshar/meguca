import {VideoItem} from "../../typings/nekotv";
import {tempNotify} from "../../ui/notification";
import {isNekoTVMuted, watchVideoDiv} from "../nekotv";
import options from "../../options";
import {IPlayer, PlayerState} from "./iplayer";

const youTubeScript = document.createElement("script");
youTubeScript.src = "https://www.youtube.com/iframe_api";


export class Youtube implements IPlayer {
    private playerEl: HTMLElement = null;
    private state = PlayerState.UNINITIALIZED;
    private videoToLoad: VideoItem|null = null;
    private ytPlayer: YT.Player;

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
        if (this.state == PlayerState.UNINITIALIZED) {
            document.head.appendChild(youTubeScript);
            console.log("Load YouTube player script");
            this.state = PlayerState.SCRIPT_LOADING;
            // @ts-ignore
            window.onYouTubeIframeAPIReady = () => {
                this.state = PlayerState.SCRIPT_LOADED;
                if(this.videoToLoad != null){
                    this.loadVideo(this.videoToLoad);
                    this.videoToLoad = null;
                }
            }
        }
    }

    public addPlayer(item:VideoItem){
        this.playerEl = document.createElement("div");
        this.playerEl.id = "youtube-player";
        watchVideoDiv.appendChild(this.playerEl);

        this.ytPlayer = new YT.Player('youtube-player', {
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
                    this.state = PlayerState.PLAYER_ADDED;
                    if (isNekoTVMuted()) {
                        this.ytPlayer.mute();
                    } else {
                        this.ytPlayer.unMute();
                    }
                    options.onChange("audioVolume", vol => {
                        this.setPlayerVolume(vol)
                    });
                    this.setPlayerVolume();
                    this.ytPlayer.playVideo();
                    console.log("player state", this.ytPlayer.getPlayerState());
                    setTimeout(() => {
                        console.log("player state", this.ytPlayer.getPlayerState());
                        if (this.ytPlayer.getPlayerState() === -1) {
                            this.ytPlayer.playVideo();
                            setTimeout(() => {
                                console.log("player state", this.ytPlayer.getPlayerState());
                                if (this.ytPlayer.getPlayerState() === -1) {
                                    tempNotify(
                                        "Click here to play this thread's live-synced video (your browser requires a click before videos can play)",
                                        "",
                                        "click-to-play",
                                        60,
                                        () => {
                                            if (this.ytPlayer && this.ytPlayer.playVideo) {
                                                this.ytPlayer.playVideo();
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

    public loadVideo(item: VideoItem)  {
        switch (this.state) {
            case PlayerState.UNINITIALIZED:
                this.initMediaPlayer();
                this.videoToLoad = item;
                break;
            case PlayerState.SCRIPT_LOADING:
                this.videoToLoad = item;
                break;
            case PlayerState.PLAYER_ADDED:
                this.ytPlayer.loadVideoById(this.extractVideoId(item.url));
                break;
            default:
                this.addPlayer(item);
        }
    }

    public isVideoLoaded(): boolean {
        return this.state == PlayerState.PLAYER_ADDED;
    }

    public play(): void {
        this.ytPlayer.playVideo();
    }

    public pause(): void {
        this.ytPlayer.pauseVideo();
    }

    public getTime(): number {
        return this.ytPlayer.getCurrentTime();
    }

    public setTime(time: number): void {
        this.ytPlayer.seekTo(time, true);
    }

    public getPlaybackRate(): number {
        return this.ytPlayer.getPlaybackRate();
    }

    public setPlaybackRate(rate: number): void {
        this.ytPlayer.setPlaybackRate(rate);
    }

    public setPlayerVolume(volume:number=null) {
        if (this.ytPlayer && this.ytPlayer.setVolume) {
            if(!volume){
                volume = options.audioVolume;
            }
            this.ytPlayer.setVolume(volume);
        }
    }

    public removeVideo() {
        switch (this.state) {
            case PlayerState.PLAYER_ADDED:
                this.ytPlayer.destroy();
                this.ytPlayer = null;
                this.playerEl.remove();
                this.playerEl = null;
                this.state = PlayerState.SCRIPT_LOADED;
                break;
            case PlayerState.SCRIPT_LOADING:
            case PlayerState.UNINITIALIZED:
                this.videoToLoad = null;
                break;
        }
    }
    public setMuted(muted: boolean) {
        if (this.ytPlayer) {
            if (muted) {
                this.ytPlayer.mute();
            } else {
                this.ytPlayer.unMute();
            }
        }
    }
}