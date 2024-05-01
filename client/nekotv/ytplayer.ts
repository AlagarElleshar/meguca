import YouTubePlayer from 'youtube-player';
import {VideoItem} from "../typings/messages";

export class Youtube {
    // private readonly player: Player;
    private readonly playerEl: HTMLElement = document.getElementById('#ytapiplayer');
    private video: HTMLElement;
    private youtube: YouTubePlayer;
    private isLoaded = false;

    // constructor(main: Main, player: Player) {
    //     this.main = main;
    //     this.player = player;
    // }

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

    public loadVideo(item: VideoItem): void {
        if (this.youtube) {
            this.youtube.loadVideoById(this.extractVideoId(item.url));
            return;
        }

        this.isLoaded = false;

        this.youtube = YouTubePlayer("watch-video", {
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
                    this.youtube.pauseVideo();
                },
                // onStateChange: (event) => {
                //     switch (event.data) {
                //         case YouTubePlayer.PlayerState.UNSTARTED:
                //             this.player.onCanBePlayed();
                //             break;
                //         case YouTubePlayer.PlayerState.ENDED:
                //         case YouTubePlayer.PlayerState.PLAYING:
                //             this.player.onPlay();
                //             break;
                //         case YouTubePlayer.PlayerState.PAUSED:
                //             this.player.onPause();
                //             break;
                //         case YouTubePlayer.PlayerState.BUFFERING:
                //             this.player.onSetTime();
                //             break;
                //     }
                // },
                // onPlaybackRateChange: () => {
                //     this.player.onRateChange();
                // },
            },
        });
        this.youtube
            // Play video is a Promise.
            // 'playVideo' is queued and will execute as soon as player is ready.
            .playVideo()
            .then(function () {
                console.log('Starting to play player1. It will take some time to buffer video before it starts playing.');
            });
    }

    public removeVideo(): void {
        if (!this.video) return;

        this.isLoaded = false;
        this.youtube.destroy();
        this.youtube = null;
        if (this.playerEl.contains(this.video)) {
            this.playerEl.removeChild(this.video);
        }
        this.video = null;
    }

    public isVideoLoaded(): boolean {
        return this.isLoaded;
    }

    public play(): void {
        this.youtube.playVideo();
    }

    public pause(): void {
        this.youtube.pauseVideo();
    }

    public getTime(): number {
        return this.youtube.getCurrentTime();
    }

    public setTime(time: number): void {
        this.youtube.seekTo(time, true);
    }

    public getPlaybackRate(): number {
        return this.youtube.getPlaybackRate();
    }

    public setPlaybackRate(rate: number): void {
        this.youtube.setPlaybackRate(rate);
    }
}