import options from "../../options";
import {IPlayer} from "./iplayer";
import {VideoItem} from "../../typings/messages";
import {vidEl} from "../nekotv";

export class RawPlayer implements IPlayer {
    private videoElement: HTMLVideoElement = null;
    getPlaybackRate(): number {
        return 0;
    }

    getTime(): number {
        return 0;
    }

    initMediaPlayer(): void {
        if(this.videoElement == null){
            this.videoElement = document.createElement('video');
            this.videoElement.id = 'raw-player';
            vidEl.appendChild(this.videoElement);
        }
    }

    isSupportedLink(url: string): boolean {
        return false;
    }

    isVideoLoaded(): boolean {
        return false;
    }

    loadVideo(item: VideoItem): void {
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
        this.videoElement.remove();
    }

}