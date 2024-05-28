import options from "../../options";
import {IPlayer} from "./iplayer";
import {VideoItem} from "../../typings/messages";
import {watchVideoDiv} from "../nekotv";

export class RawPlayer implements IPlayer {
    protected videoElement: HTMLVideoElement = null;
    protected loaded : boolean = false;
    getPlaybackRate(): number {
        return this.videoElement.playbackRate
    }

    getTime(): number {
        return this.videoElement.currentTime;
    }

    initMediaPlayer(): void {
        if(this.videoElement == null){
            this.videoElement = document.createElement('video');
            this.videoElement.id = 'raw-player';
            this.loaded = false
            const rp = this
            this.videoElement.addEventListener('loadeddata', function() {
                rp.loaded = true;
            }, false);
            watchVideoDiv.appendChild(this.videoElement);
        }
    }

    isSupportedLink(url: string): boolean {
        return false;
    }

    isVideoLoaded(): boolean {
        return this.loaded;
    }

    loadVideo(item: VideoItem): void {
        this.loaded = false;
        if(this.videoElement == null) {
            this.initMediaPlayer()
        }
        this.videoElement.src = item.url;
    }

    pause(): void {
        this.videoElement.pause()
    }

    play(): void {
        this.videoElement.play()
    }

    setMuted(isMuted: boolean): void {
        this.videoElement.muted = isMuted;
    }

    setPlaybackRate(rate: number): void {
        this.videoElement.playbackRate = rate;
    }

    setTime(time: number): void {
        this.videoElement.currentTime = time;
    }

    removeVideo(): void {
        this.loaded = false;
        if(this.videoElement) {
            this.videoElement.remove();
            this.videoElement = null;
        }
    }

}