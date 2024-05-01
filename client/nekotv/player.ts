import {VideoList} from "./videolist";
import options from "../options";
import {tempNotify} from "../ui/notification";

export class Player {
    public videoList: VideoList = new VideoList();
    private ytPlayer : YT.Player;
    private isYouTubeReady: boolean;

    private matchEmbed = /youtube\.com\/embed\/([A-z0-9_-]+)/;
    private matchShorts = /youtube\.com\/shorts\/([A-z0-9_-]+)/;
    private matchShort = /youtu\.be\/([A-z0-9_-]+)/;
    private matchId = /youtube\.com.*v=([A-z0-9_-]+)/;

    private extractVideoId(url: string): string {
        if (this.matchId.test(url)) {
            return url.match(this.matchId)[1];
        }
        if (this.matchShort.test(url)) {
            return url.match(this.matchShort)[1];
        }
        if (this.matchShorts.test(url)) {
            return url.match(this.matchShorts)[1];
        }
        if (this.matchEmbed.test(url)) {
            return url.match(this.matchEmbed)[1];
        }
        return "";
    }

    public setNextItem(pos:number) {
        this.videoList.setNextItem(pos);
    }
    public setVideo(i:number){
        let item = this.videoList.getItem(i);
        this.videoList.setPos(i);

    }
    public initYouTubePlayer() {
        console.log("Initialize YouTube player");
        this.ytPlayer = new YT.Player("watch-video", {
            playerVars: {
                enablejsapi: 1,
                iv_load_policy: 3,
                playsinline: 1,
                controls: 0,
                disablekb: 1,
                fs: 0,
                modestbranding: 1,
                rel: 0,
                showinfo: 0,
                origin: location.origin,
            },
            events: {
                onReady: this.onPlayerReady,
                onStateChange: this.onPlayerStateChange,
            },
        });
        (window as any).ytplayer = this.ytPlayer;
    }
    public onPlayerReady() {
        this.setPlayerVolume();
        // if (state.muted) {
        //     ytPlayer.mute();
        // }

        this.isYouTubeReady = true;

        const videoId = this.videoList.currentItem?.url ? this.extractVideoId(this.videoList.currentItem.url) : "";
        if (!videoId) {
            console.log("Init - no video to play");
            return;
        }

        const currentTime = this.ytPlayer.getCurrentTime();
        if (currentTime >= 1000) {
            const seekTime = currentTime / 1000 + 1;
            console.log("Load time", seekTime);
            this.ytPlayer.loadVideoById(videoId, seekTime);
        } else {
            this.ytPlayer.loadVideoById(videoId);
        }

        setTimeout(() => {
            if (this.ytPlayer.getPlayerState() === -1) {
                this.ytPlayer.playVideo();
                setTimeout(() => {
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
    }
    public setPlayerVolume() {
        if (this.ytPlayer && this.ytPlayer.setVolume) {
            this.ytPlayer.setVolume(options.watchVolume);
        }
    }
    public onSync(currentTime: number){

        const playerTime = this.ytPlayer.getCurrentTime();
        const timeDifference = Math.abs(playerTime * 1000 - currentTime);

        if (currentTime < 500 && timeDifference < 1600) {
            console.error("Jankiness detected");
        } else if (timeDifference > 1600) {
            let seekTime = Math.floor((currentTime + 100) / 1000);
            if (Math.floor(this.ytPlayer.getCurrentTime()) === seekTime) {
                console.log("Seek time incremented");
                seekTime += 1;
            }
            this.ytPlayer.seekTo(seekTime, true);
            console.log("Seek time", seekTime);
            console.log("Time out of sync - seeking");
        }
    }

    private onPlayerStateChange() {

    }
    public removeItem(url:string) {
        var index = this.videoList.findIndex(item => item.url == url);

        if (index == -1) return;

        let isCurrent = this.videoList.currentItem.url == url;
        this.videoList.removeItem(index);

        if (isCurrent && this.videoList.length > 0) {
            this.setVideo(this.videoList.pos);
        }
    }

    skipVideo(url: string) {
        var pos = this.videoList.findIndex(function(item) {
            return item.url == url;
        });
        if(pos == -1) {
            return;
        }
        this.videoList.setPos(pos);
        if(this.videoList.items.length == 0) {
            return;
        }
        this.setVideo(this.videoList.pos);
    }
}
