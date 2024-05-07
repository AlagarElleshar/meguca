import {Youtube} from "./players/youtube";
import {VideoList} from "./videolist";
import {VideoItem, VideoType} from "../typings/messages";
import {IPlayer} from "./players/iplayer";
import {TwitchPlayer} from "./players/twitch";
import {RawPlayer} from "./players/rawplayer";
import {IFramePlayer} from "./players/iframeplayer";

export class Player {
    private player: IPlayer = null;
    readonly players : Record<VideoType,IPlayer> = {
        [VideoType.IFRAME]: new IFramePlayer(),
        [VideoType.YOUTUBE]: new Youtube(),
        [VideoType.TWITCH]: new TwitchPlayer(),
        [VideoType.RAW]: new RawPlayer()
    }
    private isLoaded = false;
    private skipSetTime = false;
    private skipSetRate = false;
    public videoList = new VideoList();

    public setNextItem(pos: number): void {
        this.videoList.setNextItem(pos);
    }

    // public toggleItemType(pos: number): void {
    //     this.videoList.toggleItemType(pos);
    // }

    // private setPlayer(newPlayer: IPlayer): void {
    //     if (this.player !== newPlayer) {
    //         if (this.player !== null) {
    //             JsApi.fireVideoRemoveEvents(this.videoList.currentItem);
    //             this.player.removeVideo();
    //         }
    //     }
    //     this.player = newPlayer;
    // }

    // public getVideoData(data: VideoDataRequest, callback: (data: VideoData) => void): void {
    //     let player = this.players.find(player => player.isSupportedLink(data.url));
    //     player = player ?? this.rawPlayer;
    //     player.getVideoData(data, callback);
    // }
    //
    // public isRawPlayerLink(url: string): boolean {
    //     return !this.players.some(player => player.isSupportedLink(url));
    // }

    // public getIframeData(data: VideoDataRequest, callback: (data: VideoData) => void): void {
    //     this.iframePlayer.getVideoData(data, callback);
    // }

    public setVideo(i: number): void {
        const item = this.videoList.getItem(i);

        this.videoList.setPos(i);
        this.isLoaded = false;

        let matchedPlayer : IPlayer = this.players[item.type];
        if(matchedPlayer != this.player){
            if(this.player !== null) {
                this.player.removeVideo();
            }
            this.player = matchedPlayer;
        }
        this.player.loadVideo(item);

        // else {
        //     this.onCanBePlayed();
        // }

        // JsApi.fireVideoChangeEvents(item);
    }

    // public changeVideoSrc(src: string): void {
    //     if (this.player === null) return;
    //     const item = this.videoList.currentItem;
    //     if (item === undefined) return;
    //
    //     this.player.loadVideo({
    //         url: src,
    //         title: item.title,
    //         author: item.author,
    //         duration: item.duration,
    //         subs: item.subs,
    //         isTemp: item.isTemp,
    //         isIframe: item.isIframe
    //     });
    // }

    public removeVideo(): void {
        // JsApi.fireVideoRemoveEvents(this.videoList.currentItem);
        if (this.player !== null) {
            this.player.removeVideo();
        }
    }


    public addVideoItem(item: VideoItem, atEnd: boolean): void {
        this.videoList.addItem(item, atEnd);
    }

    public removeItem(url: string): void {
        const index = this.videoList.findIndex(item => item.url === url);
        if (index === -1) return;

        const isCurrent = this.videoList.currentItem?.url === url;
        this.videoList.removeItem(index);

        if (isCurrent && this.videoList.length > 0) {
            this.setVideo(this.videoList.pos);
        }
    }

    public skipItem(url: string): void {
        const pos = this.videoList.findIndex(item => item.url === url);
        if (pos === -1) return;

        this.videoList.setPos(pos);
        this.videoList.skipItem();

        if (this.videoList.length === 0) return;
        this.setVideo(this.videoList.pos);
    }

    public getItems(): VideoItem[] {
        return this.videoList.getItems();
    }

    public setItems(list: VideoItem[], pos?: number): void {
        const currentUrl = this.videoList.pos >= this.videoList.length ? '' : this.videoList.currentItem?.url;
        this.clearItems();

        if (list.length === 0) return;
        for (const video of list) {
            this.addVideoItem(video, true);
        }

        if (pos !== undefined) {
            this.videoList.setPos(pos);
        }

        if (currentUrl !== this.videoList.currentItem?.url || this.player === null) {
            this.setVideo(this.videoList.pos);
        }
    }

    public clearItems(): void {
        this.videoList.clear();
    }

    public refresh(): void {
        if (this.videoList.length === 0) return;
        const time = this.getTime();
        this.removeVideo();
        this.setVideo(this.videoList.pos);
    }

    public isListEmpty(): boolean {
        return this.videoList.length === 0;
    }

    public itemsLength(): number {
        return this.videoList.length;
    }

    public getItemPos(): number {
        return this.videoList.pos;
    }

    public hasVideo(): boolean {
        return this.player !== null;
    }

    public getDuration(): number {
        if (this.videoList.pos >= this.videoList.length) return 0;
        return this.videoList.currentItem?.duration ?? 0;
    }

    public isVideoLoaded(): boolean {
        return this.player?.isVideoLoaded() ?? false;
    }

    public play(): void {
        if (this.player === null) return;
        if (!this.player.isVideoLoaded()) return;
        this.player.play();
    }

    public pause(): void {
        if (this.player === null) return;
        if (!this.player.isVideoLoaded()) return;
        this.player.pause();
    }

    public getTime(): number {
        if (this.player === null) return 0;
        if (!this.player.isVideoLoaded()) return 0;
        return this.player.getTime();
    }

    public setTime(time: number, isLocal = true): void {
        if (this.player === null) return;
        if (!this.player.isVideoLoaded()) return;
        this.skipSetTime = isLocal;
        this.player.setTime(time);
    }

    public getPlaybackRate(): number {
        if (this.player === null) return 1;
        if (!this.player.isVideoLoaded()) return 1;
        return this.player.getPlaybackRate();
    }

    public setPlaybackRate(rate: number, isLocal = true): void {
        if (this.player === null) return;
        if (!this.player.isVideoLoaded()) return;
        this.skipSetRate = isLocal;
        this.player.setPlaybackRate(rate);
    }

    public setMuted(isMuted: boolean) {
        this.player.setMuted(isMuted);
    }

    public stop() {
        if(this.player) {
            this.player.removeVideo()
            this.player = null
        }
    }

    public reload() {
        this.player.removeVideo()
        this.player.loadVideo(this.videoList.currentItem)
    }
}