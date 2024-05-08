import {IPlayer} from "./iplayer";
import {VideoItem} from "../../typings/messages";
import {watchVideoDiv} from "../nekotv";

const iframeElement: HTMLIFrameElement = document.createElement('iframe');
iframeElement.id = 'youtube-player';
iframeElement.frameBorder = '0';
iframeElement.allowFullscreen = true;
iframeElement.allow = 'accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture; web-share';
iframeElement.referrerPolicy = 'strict-origin-when-cross-origin';
iframeElement.width = '640';
iframeElement.height = '360';
iframeElement.classList.add("iframe-player")

export class IFramePlayer implements IPlayer {
    private serverTime = 0;
    private currentIframe: HTMLIFrameElement = null;
    private loaded = false;

    getPlaybackRate(): number {
        return 1;
    }

    public getTime(): number {
        if (this.serverTime == 0) {
            return 0;
        }
        return Date.now() / 1000 - this.serverTime;
    }

    initMediaPlayer(): void {
    }

    isSupportedLink(url: string): boolean {
        return false;
    }

    isVideoLoaded(): boolean {
        return this.loaded;
    }

    loadVideo(item: VideoItem): void {
        if(this.currentIframe === null){
            this.currentIframe = iframeElement.cloneNode() as HTMLIFrameElement;
            this.currentIframe.src = item.id;
            this.currentIframe.title = item.title;
            this.currentIframe.onload = () => {
                this.loaded = true;
            }
            watchVideoDiv.appendChild(this.currentIframe);
        } else {
            this.currentIframe.src = item.id;
            this.currentIframe.title = item.title;
        }
    }

    pause(): void {
    }

    play(): void {
    }

    removeVideo(): void {
        if (this.currentIframe === null) return;
        this.currentIframe.remove();
        this.currentIframe = null;
        this.loaded = false;
    }

    setMuted(isMuted: boolean): void {
    }

    setPlaybackRate(rate: number): void {
    }

    public setTime(time: number): void {
        this.serverTime = Date.now() / 1000 - time;
    }
}