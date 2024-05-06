import { VideoItem } from "../../typings/messages";
import { tempNotify } from "../../ui/notification";
import {isNekoTVMuted, isNekoTVOpen, vidEl, watchMuteButton, watchPlaylistButton} from "../nekotv";
import options from "../../options";
import {IPlayer, PlayerState} from "./iplayer";
import PlayerOptions = Twitch.PlayerOptions;


const twitchRegex = /(?:https?:\/\/)?(?:www\.)?twitch\.tv\/(\w+)(?:\/)?/;
export class TwitchPlayer implements IPlayer {
    private readonly playerEl: HTMLElement = document.getElementById('#twitchplayer');
    private videoToLoad: VideoItem | null = null;
    private twitchPlayer : Twitch.Player;
    private twitchPlayerDiv : HTMLElement;
    private state = PlayerState.UNINITIALIZED;
    private serverTime = 0;

    public isSupportedLink(url: string): boolean {
        return url.match(twitchRegex) != null;
    }

    public extractChannelName(url: string): string {
        const match = url.match(twitchRegex);
        return match ? match[1] : null;
    }

    public initMediaPlayer() {
        if (this.state == PlayerState.UNINITIALIZED){
            const script = document.createElement('script');
            script.setAttribute('src', 'https://player.twitch.tv/js/embed/v1.js');
            document.head.appendChild(script);
            console.log("Load Twitch player script");
            this.state = PlayerState.SCRIPT_LOADING;

            script.onload = () => {
                this.state = PlayerState.SCRIPT_LOADED;
                if (this.videoToLoad != null) {
                    this.loadVideo(this.videoToLoad);
                    this.videoToLoad = null;
                }
            };
        }
    }

    public addPlayer(item: VideoItem) {
        const channelName = this.extractChannelName(item.url);

        const twitchOptions: PlayerOptions = {
            channel: channelName,
            autoplay: true,
        };

        this.twitchPlayer = new Twitch.Player("watch-video", twitchOptions);
        this.twitchPlayer.addEventListener(Twitch.Player.READY, () => {
            watchPlaylistButton.style.display = "";
            watchMuteButton.style.display = "none";
            this.state = PlayerState.PLAYER_ADDED;
            if (isNekoTVMuted()) {
                this.twitchPlayer.setMuted(true);
            } else {
                this.twitchPlayer.setMuted(false);
            }
            options.onChange("audioVolume", (vol) => {
                this.setPlayerVolume(vol);
            });
            this.setPlayerVolume();
            this.twitchPlayer.play();
            setTimeout(() => {
                if (this.twitchPlayer.isPaused()) {
                    this.setPlayerVolume()
                    this.twitchPlayer.play();
                    setTimeout(() => {
                        if (this.twitchPlayer.isPaused()) {
                            tempNotify(
                                "Click here to play this thread's live-synced video (your browser requires a click before videos can play)",
                                "",
                                "click-to-play",
                                60,
                                () => {
                                    if (this.twitchPlayer && this.twitchPlayer.play) {
                                        this.setPlayerVolume()
                                        this.twitchPlayer.play()
                                    }
                                }
                            );
                        }
                    }, 1000);
                }
            }, 2000);
        });
    }

    public loadVideo(item: VideoItem) {
        switch (this.state) {
            case PlayerState.UNINITIALIZED:
                this.initMediaPlayer()
                this.videoToLoad = item;
                break;
            case PlayerState.SCRIPT_LOADING:
                this.videoToLoad = item;
                break;
            case PlayerState.SCRIPT_LOADED:
                this.addPlayer(item);
                break;
            case PlayerState.PLAYER_ADDED:
                this.twitchPlayer.setVideo(this.extractChannelName(item.url));
                break;
        }
    }

    public removeVideo(): void {
        if (!this.twitchPlayer) return;
        this.twitchPlayer = null;
        let twitchIframe = vidEl.querySelector(`iframe[title="Twitch"]`)
        twitchIframe.remove()
        watchPlaylistButton.style.display = "none";
        watchMuteButton.style.display = "";
        this.state = PlayerState.SCRIPT_LOADED;
    }

    public isVideoLoaded(): boolean {
        return this.state === PlayerState.PLAYER_ADDED;
    }

    public play(): void {
        this.twitchPlayer.play();
    }

    public pause(): void {
        this.twitchPlayer.pause();
    }

    public getTime(): number {
        if(this.serverTime == 0){
            return 0;
        }
        return Date.now() / 1000 - this.serverTime;
    }

    public setTime(time: number): void {
        this.serverTime = Date.now() / 1000 - time;
    }

    public getPlaybackRate(): number {
        return 1; // Twitch player doesn't have playback rate control
    }

    public setPlaybackRate(rate: number): void {
        // Twitch player doesn't have playback rate control
    }

    public setPlayerVolume(volume: number = null) {
        if(volume == null){
            volume = options.audioVolume
        }
        if (this.twitchPlayer && this.twitchPlayer.setVolume) {
            this.twitchPlayer.setVolume(volume/100)
            this.twitchPlayer.setMuted(isNekoTVMuted());
        }
    }


    public setMuted(muted: boolean) {
        if (this.twitchPlayer) {
            this.twitchPlayer.setMuted(muted);
        }
    }
}