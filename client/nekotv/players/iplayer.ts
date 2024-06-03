import {VideoItem} from "../../typings/nekotv";

export enum PlayerState {
    UNINITIALIZED = 0,
    SCRIPT_LOADING = 1,
    SCRIPT_LOADED = 2,
    PLAYER_ADDED = 3,
}
export interface IPlayer {
    isSupportedLink(url: string): boolean;
    loadVideo(item: VideoItem): void;
    isVideoLoaded(): boolean;
    play(): void;
    pause(): void;
    getTime(): number;
    setTime(time: number): void;
    getPlaybackRate(): number;
    setPlaybackRate(rate: number): void;
    setMuted(isMuted: boolean): void;
    initMediaPlayer(): void;
    removeVideo(): void;
}